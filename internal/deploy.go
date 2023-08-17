// Package internal contains all logic for deployment service
package internal

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zos/pkg/gridtypes"
)

func (d *Deployer) Deploy(ctx context.Context) error {
	statusUp := "up"
	falseVal := true
	trueVal := true
	oneVal := uint64(1)
	memoryGB := uint64(8)
	diskGB := uint64(50)
	farmID := uint64(1)
	minRootfs := *convertGBToBytes(2)

	log.Debug().Str("mnemonics", d.configs.mnemonic).Str("network", d.configs.network).Msg("Initializing threefold plugin client...")
	tfPluginClient, err := deployer.NewTFPluginClient(d.configs.mnemonic, "sr25519", d.configs.network, "", "", "", 0, false)
	if err != nil {
		return err
	}

	nodeFilter := types.NodeFilter{
		Status:  &statusUp,
		FreeSRU: convertGBToBytes(diskGB * 2),
		FreeMRU: convertGBToBytes(memoryGB),
		FarmIDs: []uint64{farmID},
		Rented:  &falseVal,
		FreeIPs: &oneVal,
		IPv4:    &trueVal,
	}

	log.Debug().Msg("Filtering nodes")
	nodes, err := deployer.FilterNodes(ctx, tfPluginClient, nodeFilter, []uint64{*convertGBToBytes(diskGB), *convertGBToBytes(diskGB)}, nil, []uint64{minRootfs})
	if err != nil {
		return err
	}

	nodeID := uint32(nodes[0].NodeID)
	log.Debug().Uint32("node ID", nodeID).Msg("Node is found")

	net := workloads.ZNet{
		Name:        fmt.Sprintf("network_%s", d.configs.vmName),
		Description: "network for deployment",
		Nodes:       []uint32{nodeID},
		IPRange: gridtypes.NewIPNet(net.IPNet{
			IP:   net.IPv4(10, 20, 0, 0),
			Mask: net.CIDRMask(16, 32),
		}),
		AddWGAccess: false,
	}

	dataDisk := workloads.Disk{
		Name:   "data_disk",
		SizeGB: 50,
	}

	dockerDisk := workloads.Disk{
		Name:   "docker_disk",
		SizeGB: 50,
	}

	vm := workloads.VM{
		Name:       d.configs.vmName,
		Flist:      "https://hub.grid.tf/tf-official-apps/threefoldtech-ubuntu-22.04.flist",
		CPU:        4,
		PublicIP:   true,
		Planetary:  true,
		Memory:     8 * 1024,
		Entrypoint: "/sbin/zinit init",
		EnvVars: map[string]string{
			"SSH_KEY": d.configs.sshKey,
		},
		Mounts: []workloads.Mount{
			{DiskName: dataDisk.Name, MountPoint: "/mydata"},
			{DiskName: dockerDisk.Name, MountPoint: "/var/lib/docker"},
		},
		NetworkName: net.Name,
	}

	log.Debug().Str("Network", net.Name).Msg("Deploying network")
	err = tfPluginClient.NetworkDeployer.Deploy(ctx, &net)
	if err != nil {
		return err
	}

	log.Debug().Str("VM", vm.Name).Msg("Deploying virtual machine")
	dl := workloads.NewDeployment(d.configs.vmName, nodeID, "", nil, net.Name, []workloads.Disk{dataDisk, dockerDisk}, nil, []workloads.VM{vm}, nil)
	err = tfPluginClient.DeploymentDeployer.Deploy(ctx, &dl)
	if err != nil {
		return err
	}

	log.Debug().Str("VM", d.configs.vmName).Msg("Loading virtual machine")
	outputVM, err := tfPluginClient.State.LoadVMFromGrid(nodeID, vm.Name, dl.Name)
	if err != nil {
		return err
	}

	yggIP := outputVM.YggIP
	log.Debug().Str("Yggdrasil IP", yggIP)

	installDockerCmds := `apt update -y &&
	apt install -y apt-transport-https ca-certificates curl software-properties-common &&
	curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - &&
	add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu focal stable" &&
	apt-cache policy docker-ce &&
	apt install -y docker-ce &&
	echo -e 'exec: bash -c "dockerd -D"\ntest: docker ps' >> /etc/zinit/dockerd.yaml &&
	zinit monitor dockerd`

	log.Debug().Msg("Installing docker")
	_, err = remoteRun("root", yggIP, installDockerCmds, d.configs.privateKey)
	if err != nil {
		return err
	}

	installCaddyCmds := `apt install -y debian-keyring debian-archive-keyring apt-transport-https &&
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg &&
	curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list &&
	apt update -y &&
	apt install -y caddy`

	log.Debug().Msg("Installing caddy")
	_, err = remoteRun("root", yggIP, installCaddyCmds, d.configs.privateKey)
	if err != nil {
		return err
	}

	log.Debug().Str("repository", d.configs.repoURL).Msg("Cloning the repository")
	_, err = remoteRun("root", yggIP, fmt.Sprintf("cd /mydata && git clone %s", d.configs.repoURL), d.configs.privateKey)
	if err != nil {
		return err
	}

	repoName := d.configs.repoURL[strings.LastIndex(d.configs.repoURL, "/")+1:]
	log.Debug().Str("repository name", repoName).Send()

	if len(d.configs.configFilePath) != 0 {
		log.Debug().Msg("Inserting repository d.configuration file")

		repoConfig, err := os.ReadFile(d.configs.configFilePath)
		if err != nil {
			return err
		}

		_, err = remoteRun("root", yggIP, fmt.Sprintf("cd /mydata && echo -e '%s' >> %s/%s", repoConfig, repoName, filepath.Base(d.configs.configFilePath)), d.configs.privateKey)
		if err != nil {
			return err
		}
	}

	log.Debug().Msg("Inserting caddy script")
	caddyFileContent := fmt.Sprintf(`%s {
  route /v1/* {
           uri strip_prefix /*
           reverse_proxy http://127.0.0.1:%d
        }
  route /* {
           uri strip_prefix /*
           reverse_proxy http://127.0.0.1:%d
        }
}`, outputVM.ComputedIP, d.configs.backendPort, d.configs.frontendPort)
	log.Debug().Str("Caddy file content", caddyFileContent)

	_, err = remoteRun("root", yggIP, fmt.Sprintf("cd /mydata && echo -e '%s' >> %s/Caddyfile", caddyFileContent, repoName), d.configs.privateKey)
	if err != nil {
		return err
	}

	log.Debug().Msg("Inserting caddy service into zinit")
	_, err = remoteRun("root", yggIP, fmt.Sprintf(`echo 'exec: bash -c "caddy run --d.configs=/mydata/%s/Caddyfile"' >> /etc/zinit/caddy.yaml && zinit monitor caddy`, repoName), d.configs.privateKey)
	if err != nil {
		return err
	}

	log.Debug().Str("repository service name", repoName).Msg("Inserting repository service into zinit")
	_, err = remoteRun("root", yggIP, fmt.Sprintf(`echo 'exec: bash -c "cd /mydata/%s && docker compose up --force-recreate"' >> /etc/zinit/%s.yaml && zinit monitor %s`, repoName, repoName, repoName), d.configs.privateKey)
	if err != nil {
		return err
	}

	return nil
}
