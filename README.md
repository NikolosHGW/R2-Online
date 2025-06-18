# R2 Online Server Emulator

This repository contains a C# implementation of an R2 Online server emulator. It includes packet definitions and various utilities.

## Generating Packet Documentation

A script is provided to automatically parse packet model definitions and create a Markdown document describing all client/server packets.

Run the script from the repository root:

```bash
python3 Scripts/generate_packet_docs.py
```

The generated file will be written to `docs/PacketDocumentation.md`.

