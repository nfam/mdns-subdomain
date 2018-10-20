# mdns-subdomain

`mdns-subdomain` routes all subdomains to a single host using CNAME in mDNS. Use this along with Avahi

## Usage

Following example will route `*.server.local` to `server.local`

1. Run as `--network=host`

    ```sh
    docker run -d \
        --name=mdns-subdomain \
        --restart=always \
        --network=host \
        -e IFACE=en0 \
        -e HOSTNAME=server.local \
        nfam/mdns-subdomain
    ```

2. Run with birdge network

    ```sh
    docker run -d \
        --name=mdns-subdomain \
        --restart=always \
        -p 5353:5353/udp \
        -e HOSTNAME=server.local
        nfam/mdns-subdomain
    ```

## Options

Argument     | Environement  | Description
:--          | :---          | :------
iface        | IFACE         | Specifies the network interface for the mNDS server listen to. If not provided, the server listens to all interfaces.
hostname     | HOSTNAME      | Specifies the hostname (ends with `.local`) to broadcast over the network. If not provided, the server hostname will be employed.
