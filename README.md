# Quantum List

SDLE ([Large Scale Distributed Systems](https://sigarra.up.pt/feup/en/UCURR_GERAL.FICHA_UC_VIEW?pv_ocorrencia_id=501934)) project by:

|Name|id|
|-|-|
| André Ismael Ferraz Ávila | up202006767 |
| Flávio Lobo Vaz | up201509918 |
| Diogo André Pereira Babo| up202004950 |

## How to run

This project contains several components that can be run.

### Load Balancer

To run a load balancer you can either use

`go run load_balancer/* <own_port>`

or

`make run_load_balancer OWN_PORT=<own_port>`

where `9988` is the port the load balancer will be binded to.

### Database Node

To run a database node you can either:

go into `./database_node/` and use

` go run . <own_port> localhost 9988`

or simply use

`make run_db_node OWN_PORT=<own_port> BAL_ADDR=<load_balancer_address> BAL_PORT=<load_balancer_port>`

A database node can be ran with the load balancer address and port values omitted, however, their port must have been at some point connected to a load balancer in order to be rediscovered the load balancer.

### App

To run the app you must go into `./app/` and use

`npm install && npm run tauri dev`

or

`npm install && npm run tauri build` and then run the built application present in the `app/src-tauri/target/release/bundle/` directory

In order to run multiple instances of the app on the same machine extra steps are required:

Build the app running the command `npm run tauri build` inside the **app** directory, extract the build from the `app/src-tauri/target/release/bundle/` directory, replace the name of the database file on line 18 of the file `app/src-tauri/src/database.rs` and then build again. The previous steps creates two equal applications that do not share the same database file.
