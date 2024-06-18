# StreamSculpt

[![Latest release](https://img.shields.io/github/v/release/synternet/StreamSculpt)](https://github.com/synternet/StreamSculpt/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/synternet/StreamSculpt/github-ci.yml?label=github-ci)](https://github.com/synternet/StreamSculpt/actions/workflows/github-ci.yml)

StreamSculpt lets you filter and unpacks Ethereum transaction smart contract receipt event log for given list of ABIs.
StreanSculpt uses Synternet Data Layer as Ethereum event logs source. Once event log is decoded it is pushed
to Synternet Data Layer as a new data stream.

See [Data Layer Quick Start](https://docs.synternet.com/build/) to learn more about Synternet Data Layer.

## Example

1. Ethereum log event received on "synternet.ethereum.log-event" subject:
```json
{
    "address": "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc",
    "blockHash": "0xd1cf5806065d2d69e5a48b51cda656a61b1a2472ce105794354798db523b66c2",
    "blockNumber": "0x11a7ba2",
    "data": "0x0000000000000000000000000000000000000000000000000000000006246db60000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c037bb78ad06d2",
    "logIndex": "0x10a",
    "removed": false,
    "topics": [
        "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822",
        "0x0000000000000000000000003fc91a3afd70395cd496c647d5a6cc9d4b2b7fad",
        "0x0000000000000000000000003fc91a3afd70395cd496c647d5a6cc9d4b2b7fad"
    ],
    "transactionHash": "0xed11df97399204f99bad0047bc60ea3974020c4403252890ef1fbedb28dfec76",
    "transactionIndex": "0x91"
}
```
2. Matched against existing ABIs in `./internal/service/abi/` directory.
3. Published to "synternet.ethereum.unpacked-log-event.0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc.Swap" subject:
```json
{
    "address": "0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc",
    "blockHash": "0xd1cf5806065d2d69e5a48b51cda656a61b1a2472ce105794354798db523b66c2",
    "blockNumber": "0x11a7ba2",
    "data": {
        "amount0In": 103050678,
        "amount0Out": 0,
        "amount1In": 0,
        "amount1Out": 54104473851463378
    },
    "logIndex": "0x10a",
    "removed": false,
    "sig": "Swap(address,uint256,uint256,uint256,uint256,address)",
    "topics": [
        "0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822",
        "0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD",
        "0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD"
    ],
    "transactionHash": "0xed11df97399204f99bad0047bc60ea3974020c4403252890ef1fbedb28dfec76",
     "transactionIndex": "0x91"
}
```

## Usage

1. Add to `./internal/service/abi/` and `./internal/service/abi/map/` ABIs of smart contracts to unpack.
Note: `go:embed` baked files into executable, must be added before compiling.

2. Compile code.
```
make build
```

3. Run executable.
```
./streamsculpt [flags]
```

## Defaults

### ABIs

Directory `./internal/service/abi/` contains smart contracts ABIs to filter and unpack.
Format is <smart-contract>.json, where <smart-contract> is smart contract address, e.g.: 0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852.json.
Content of file is ABI of smart contract in JSON format.

Default `./internal/service/abi/` content:
```
  abi/
        0x0d4a11d5eeaac28ec3f61d100daf4d40471f1852.json // ETH-USDT Uniswap Pool
        0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc.json // USDC-ETH Uniswap Pool
        0xb8a1a865e4405281311c5bc0f90c240498472d3e.json // NOIA-ETH Uniswap Pool
        uniswapv2pair.json // UniswapV2Pair (1)
      map/
            uniswap2pair.json // (2)
```

From example above you will notice `uniswap2pair.json` `(1)` not mentioned before. This is used together with
`./internal/service/abi/map/uniswapv2pair.json` `(2)`. `(2)` defined a list of Smart Contract addressed to be mapped
to `(1)`. This is useful when you have an ABI which is identical to more than one deployed Smart Contracts would
like to avoid repeating yourself.

### Flags

| Flag                                 | Description                                                       |
| ------------------------------------ | ----------------------------------------------------------------- |
| nats-urls                            | NATS servers URLs (comma separated)                               |
| nats-sub-nkey                        | NATS user credentials NKey string to subscribe to ETH stream      |
| nats-pub-nkey                        | NATS user credentials NKey string to publish unpacked event log   |
| nats-reconnect-wait                  | NATS reconnect wait duration                                      |
| nats-max-reconnect                   | NATS max reconnect attempts count                                 |
| nats-event-log-stream-subject        | NATS event log stream subject                                     |
| nats-unpacked-streams-subject-prefix | NATS unpacked streams prefix                                      |

- `nats-*`. NATS.
`nats-sub-nkey`, `nats-pub-nkey`, `nats-unpacked-streams-subject-prefix` must be provided. Uses Synternet Data Layer to get Ethereum transactions event log. See [Data Layer Quick Start](https://docs.synternet.com/build/data-layer/data-layer-quick-start) to learn more.

## Docker

1. Build image.
```
docker build -f ./docker/Dockerfile -t streamsculpt .
```

2. Run container with passed environment variables.
```
docker run -it --rm --env-file=.env streamsculpt
```

Note: [Flags](#flags) can be passed as environment variables.
Environment variables are all caps flags separated with underscore. See `./docker/entrypoint.sh`.

## Contributing

We welcome contributions from the community. Whether it's a bug report, a new feature, or a code fix, your input is valued and appreciated.

## Synternet

If you have any questions, ideas, or simply want to connect with us, we encourage you to reach out through any of the following channels:

- **Discord**: Join our vibrant community on Discord at [https://discord.com/invite/jqZur5S3KZ](https://discord.com/invite/jqZur5S3KZ). Engage in discussions, seek assistance, and collaborate with like-minded individuals.
- **Telegram**: Connect with us on Telegram at [https://t.me/Synternet](https://t.me/Synternet). Stay updated with the latest news, announcements, and interact with our team members and community.
- **Email**: If you prefer email communication, feel free to reach out to us at devrel@synternet.com. We're here to address your inquiries, provide support, and explore collaboration opportunities.
