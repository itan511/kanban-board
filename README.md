# Kanban board

Kanban board in Golang and PostgreSQL

## Installation

To use the application you need to install [Docker](https://www.docker.com/get-started/)

After installation, check that Docker is running using the command:
```bash
docker --version
```

## Usage

To run the application you can clone the repository
```bash
git clone https://github.com/itan511/kanban-board.git
```

To run the application, you need to create a .env file in the project root folder 
and add environment variables there as in example.env.
After that, go to the terminal in the project directory and enter the following commands
To run application enter:
```Makefile
make all
```

To build database container enter:
```Makefile
make build-db
```

To build application container enter:
```Makefile
make build-app
```

To run database container enter:
```Makefile
make run-db
```

To run application container enter:
```Makefile
make run-app
```

To stop containers enter:
```Makefile
make stop
```

To remove containers and volumes enter:
```Makefile
make clean
```
