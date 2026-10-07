# goswarm

Tiny BitTorrent client written in Go, built from scratch.

## Install

```bash
git clone https://github.com/halwaii/goswarm.git
cd goswarm
go build
```

## Usage

Try downloading Debian!

```bash
goswarm debian-amd64-netinst.iso.torrent debian.iso
```

Or run it directly without building:

```bash
go run . debian-amd64-netinst.iso.torrent debian.iso
```

## Demo



## Features

- Parses `.torrent` files (bencode)
- Announces to the tracker and fetches peers
- Downloads from multiple peers concurrently
- Verifies every piece with SHA-1 before saving

## Limitations

- Only supports `.torrent` files (no magnet links)
- Only supports HTTP trackers (no UDP)
- Does not support multi-file torrents
- Strictly leeches (does not support uploading pieces)
