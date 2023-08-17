# App_tfgrid_deployer

<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-0%25-brightgreen.svg?longCache=true&style=flat)</a> [![Testing](https://github.com/rawdaGastan/app_tfgrid_deployer/actions/workflows/test.yml/badge.svg?branch=development)](https://github.com/rawdaGastan/app_tfgrid_deployer/actions/workflows/test.yml) [![Testing](https://github.com/rawdaGastan/app_tfgrid_deployer/actions/workflows/lint.yml/badge.svg?branch=development)](https://github.com/rawdaGastan/app_tfgrid_deployer/actions/workflows/lint.yml)

A binary to simplify deploying apps using threefold grid

## Create a Threefold account

- You should have an account linked to a network to be able to use the deployer.
- If you don't have an account, [create](https://threefoldtech.github.io/info_grid/dashboard/portal/dashboard_portal_polkadot_create_account.html) a new account.

## Usage

- You should have an threefold grid [account](#create-a-threefold-account).
- Add your [configurations](#configuration).
- [Build](#build) your app_tfgrid_deployer binary.
- Start deploying:

```bash
app_tfgrid_deployer deploy -c <config-file-path>
```

- You can update deployment:

```bash
app_tfgrid_deployer update -c <config-file-path> -y <your vm yggdrasil IP>
```

Make sure your repo has:

- `docker compose file`, `backend docker file` and `frontend docker file`
- 2 directories for `backend` and `frontend`
- your docker compose images for backend and frontend have the same name as their directories.

## Build

- Build your binary
- Move the binary to any of `$PATH` directories, for example:

```bash
make build
sudo mv bin/app_tfgrid_deployer /usr/local/bin
```

## Configuration

Before building, you need to add your configurations.

example `.env`:

```env
MNEMONIC="<your mnemonic here>"
NETWORK="<the network you want to use (dev, qa, test, main)>"
VM_NAME="<your vm name>"
REPO_URL="<your repo url>"
CONFIG_FILE_NAME="<your config file path that will be inserted in the repo>"
BACKEND_DIR="<your backend directory name, make sure it has the same name as the backend docker image>"
FRONTEND_DIR="<your frontend directory name, make sure it has the same name as the frontend docker image>"
BACKEND_PORT="<your backend port>"
FRONTEND_PORT="<your frontend port>"
```

## Test

```bash
make test
```
