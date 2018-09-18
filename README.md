# Enclave (Server)
See [enclave-app](https://github.com/Vauxel/enclave-app) for more information.

## Installing and Running an Instance

Clone the repository to your machine
```
git clone https://github.com/Vauxel/enclave-server.git
cd enclave-server
```

Complete [the instructions to compile libsodium](https://github.com/GoKillers/libsodium-go#how-to-build)

_if libsodium's build.sh fails, export the correct pkg_config path_
```
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
```

_if libsodium.so.* cannot be found on runtime, link it from the local/lib directory, replacing * with the version number_
```
ln -s /usr/local/lib/libsodium.so.* /usr/lib/libsodium.so.*
```

Download and install the other dependencies
```
go get ./...
```

Compile the source
```
go build
```

Run the executable
```
./enclave-server
```

Connect using the [client](https://github.com/Vauxel/enclave-app)!
