# GO-POW
## Word of Wisdom TCP Server

This project implements a TCP server that provides quotes from the "Word of Wisdom" book or any other collection of quotes. The server is protected from DDOS attacks using a Proof of Work (POW) challenge-response protocol based on the Hashcash protocol.

## Table of Contents

- [Project Structure](#project-structure)
  - [Server](#server)
  - [Client](#client)
  - [Environment variables](#environment-variables)
- [Proof of Work Algorithm](#proof-of-work-algorithm)
- [Communication protocol](#communication-protocol)
- [Package hashcash](#package-hashcash)
  - [Implementation](#implementation)
  - [Hashcash header](#hashcash-header)
  - [SHA-1 Hash Function](#sha-1-hash-function)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [Testing](#testing)

## Project Structure
### Server
This component implements the TCP server. It verifies the Proof of Work and responds with a quote upon successful verification.
- **cmd/main.go**: entry point of the server application.
- **internal**:
    - **server**: contains logic for handling connections and communication with clients.
- **pkg**:
    - **book**: provides functionality to parse and retrieve quotes from a source.
    - **logger**: offers a customizable logger with the ability to set the log level to DEBUG.
    - **config**: provides methods for parsing a configuration file (`conf.yaml`) to obtain project settings.
    - **hashcash**: implements parsing and validation of Hashcash headers for proof-of-work.
    - **random**: provides functionality to generate a random portion for the Hashcash header.
- **source**: contains the source file (`quotes.txt`) used for parsing quotes in the book.

### Client
This component solves the Proof of Work challenge and sends a request to the server for a quote.
- **cmd/main.go**: entry point of the client application.
- **internal**:
    - **client**: contains logic for handling connections and communication with clients.
- **pkg**:
    - **logger**: offers a customizable logger with the ability to set the log level to DEBUG.
    - **config**: provides methods for parsing a configuration file (`conf.yaml`) to obtain project settings.
    - **hashcash**: implements parsing and validation of Hashcash headers for proof-of-work.

### Environment variables

### Server

| name             | type    | default        | description
|------------------|---------|----------------|-----------------------------------------------
| SERVER_HOST      | string  | "localhost"    | server TCP host. Default is set in `conf.yaml`
| SERVER_PORT      | string  | ":40999"       | server TCP host. Default is set in `conf.yaml`
| DEBUG            | bool    | false          | set debug level for a logger

### Client

| name           | type    | default        | description
|----------------|---------|----------------|-----------------------------------------------
| SERVER_HOST    | string  | "localhost"    | listen TCP host. Default is set in `conf.yaml`
| SERVER_PORT    | string  | ":40999"       | listen TCP host. Default is set in `conf.yaml`
| DEBUG          | bool    | false          | set debug level for a logger



## Proof of Work Algorithm
The chosen Proof of Work algorithm is based on the Hashcash protocol, which uses the SHA-1 cryptographic hash function. This algorithm requires computational effort to find the correct nonce that satisfies the specified difficulty level.

### Hashcash advantages:
- Proven Effectiveness: Hashcash has a track record of success, particularly in mitigating email spam and preventing denial-of-service attacks. It has been used in real-world applications for these purposes.

- Simplicity: Hashcash is relatively simple to implement. This makes it accessible for developers who may not have extensive experience with cryptographic algorithms.

- Abundance of Documentation: There is a wealth of documentation and articles available that explain Hashcash in detail. This can be valuable for understanding its inner workings and effectively implementing it.

- Server-Side Validation: Verifying Hashcash on the server side is straightforward. This simplifies the process of validating proof of work, ensuring that computational effort has been expended.

- Dynamic Complexity Management: Hashcash allows you to dynamically manage the complexity required for clients to solve the challenge. This means you can adapt the difficulty level based on the capabilities of the client's machine.

- Resilience Against Pre-Computation: While there is a concern about clients pre-computing challenges for potential DDOS attacks, this can be mitigated by implementing additional validation measures on the server. For example, the server can use caching and checks to ensure that challenges are generated and verified in real-time.

### Hashcash disadvantages:
- Compute Time Dependency on Client's Machine Power. The time taken to compute Hashcash depends on the computational power of the client's machine. This can lead to variability in the time it takes to solve the challenge, potentially causing issues for clients with very weak computing resources.

- Potential for Powerful Computers to Implement DDOS Attacks.Very powerful computers may have the capability to compute Hashcash challenges quickly, which could potentially be exploited for launching Distributed Denial-of-Service (DDoS) attacks. This risk arises from the computational power imbalance between different clients.

- Dynamic Complexity Management Can Be Tricky. While the ability to dynamically manage the complexity required for clients is a feature of Hashcash, it can be challenging to strike the right balance. Setting the difficulty level too high may exclude legitimate users with limited computational resources, while setting it too low may make it susceptible to abuse.

## Communication protocol

1. Client establishes a connection with the server.
2. Server initiates the protocol by sending a challenge header to the client.
3. Client parse a challenge header and calculate a solution .
4. Client send a solution header to the server.
5. Server parse and validate a solution header.
6. If a solution header is valid:

    - Server sends a quote from “Word of Wisdom” to the client.
    - Client logs the response from the server.
9. Connection close.

## Package hashcash
### Implementation
The `hashcash` package is included in both the `client` and `server` directories for convenient initial development. However, it is recommended to eventually separate this package into its own repository and import it as a library for more organized and modular code.

This package is implemented based on the information available in the [Hashcash Wikipedia page](https://en.wikipedia.org/wiki/Hashcash).


### Hashcash header
The header line looks something like this:

```1:20:1303030600:255.255.0.0:80::McMybZIhxKXu57jd:ckvi```

The header contains:
- **ver**: Hashcash format version, 1 (which supersedes version 0).
- **bits**: Number of "partial pre-image" (zero) bits in the hashed code.
- **date**: The time that the message was sent, in the format `YYMMDD[hhmm[ss]]`.
- **resource**: Resource data string being transmitted an IP address.
- **rand**: String of random characters, encoded in base-64 format.
- **counter**: Binary counter, encoded in base-64 format.

The package includes a regular expression for parsing and validating the header:

```^(\d):(\d+):(\d+):(\d+\.\d+\.\d+\.\d+:\d+)::([A-Za-z0-9+/]+={0,2}):([A-Za-z0-9+/]+={0,2})$```

### SHA-1 Hash Function
The package utilizes the SHA-1 (Secure Hash Algorithm 1) hash function for generating and validating Hashcash headers. SHA-1 is a widely used cryptographic hash function that produces a 160-bit (20-byte) hash value. It is known for its simplicity, which, in combination with the customizable `zeroBits` parameter, allows for increased computational work in hash calculation.

By adjusting `zeroBits` in `conf.yaml`, users can increase the computational effort required to find a valid Hashcash header, enhancing the security of the proof-of-work process.

## Getting Started

### Prerequisites

- [Golang](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Installation

1. Clone the repository:

```bash
git clone git@github.com:istomin10593/go-pow.git
cd go-pow
```
2. Build Docker images and launch both the server and client in a Docker environment:
```
make up
```

3. Stop the server and client:

```
make down
```

4. Run the server on local machine:

```
make server-run
```

5. Run client on local machine:

```
make client-run
```

## Usage 
After executing the project with the command: `make up`, the console will display logs similar to the following, confirming a successful launch. The file `client/conf.yaml` allows for configuration of the number of clients `client.number` and their intervals `client.timeout`.

```
go-pow-git-server-1  | 2023-09-20T14:47:26.714Z INFO    server/server.go:43     listening       {"port": ":40999"}
go-pow-git-client-1  | 2023-09-20T14:47:27.033Z INFO    client/client.go:87     connected to    {"clientID": 0, "address": "server:40999"}
go-pow-git-server-1  | 2023-09-20T14:47:27.034Z INFO    server/server.go:144    pow validation successful       {"resource": "172.28.0.3:51844"}
go-pow-git-server-1  | 2023-09-20T14:47:27.035Z INFO    server/server.go:72     handleConnection succeeded      {"client address": "172.28.0.3:51844"}
go-pow-git-client-1  | 2023-09-20T14:47:27.035Z INFO    client/client.go:140    received response from server   {"clientID": 0, "response": "The successful warrior is the average man, with laser-like focus.~Bruce Lee"}
go-pow-git-client-1  | 2023-09-20T14:47:27.533Z INFO    app/main.go:60  client shutting down...
go-pow-git-client-1  | 2023-09-20T14:47:27.533Z INFO    client/client.go:57     pow completed successfully      {"clientID": 0}
go-pow-git-client-1  | 2023-09-20T14:47:27.533Z INFO    app/main.go:67  successful shutdown
```
## Testing
Tests cover the logic of the `hashcash package`, as well as the connection handling logic in both the `client` and `server` applications, including integration and unit tests.
To run the tests  use the following command:
```
make test
```