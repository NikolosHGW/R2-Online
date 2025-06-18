import re
from pathlib import Path

PACKET_ENUM_PATH = Path('Packets/Packets.Core/Enums/PacketType.cs')
MODELS_ROOT = Path('Packets')
OUTPUT_PATH = Path('docs/PacketDocumentation.md')


def parse_packet_enum():
    text = PACKET_ENUM_PATH.read_text(encoding='utf-8', errors='ignore')
    pairs = re.findall(r"(\w+)\s*=\s*(\d+)", text)
    return {name: int(value) for name, value in pairs}


def parse_model(file: Path, enum_map):
    text = file.read_text(encoding='utf-8', errors='ignore')
    direction = 'Server -> Client' if '/Send/' in file.as_posix() else 'Client -> Server'
    summary_match = re.search(r"<summary>\s*(.*?)\s*</summary>", text, re.S)
    if summary_match:
        raw = summary_match.group(1)
        raw = re.sub(r"\s*///\s*", "", raw)
        summary = ' '.join(raw.split())
    else:
        summary = 'No description.'
    attr_match = re.search(r"\[Model\s*\(\s*(?:[A-Za-z0-9_.]*\.)?PacketType\.([A-Za-z0-9_]+)\s*\)\s*\]", text)
    packet_enum = attr_match.group(1) if attr_match else 'Unknown'
    packet_code = enum_map.get(packet_enum)
    fields = re.findall(r"public\s+([A-Za-z0-9_<>,\[\]]+)\s+([A-Za-z0-9_]+)\s*{\s*get;", text)
    return {
        'file': file,
        'class_name': file.stem,
        'packet_enum': packet_enum,
        'packet_code': packet_code,
        'direction': direction,
        'description': summary,
        'fields': fields,
    }


def collect_models(enum_map):
    # search for all model definitions in send/receive folders
    files = list(Path('Packets').glob('Packets.Server.*/**/Models/**/*.cs'))
    model_files = []
    for f in files:
        text = f.read_text(encoding='utf-8', errors='ignore')
        if '[Model(' in text:
            model_files.append(f)
    models = [parse_model(f, enum_map) for f in model_files]
    models.sort(key=lambda m: (m['packet_code'] if m['packet_code'] is not None else 0))
    return models


def generate_docs(models):
    lines = ["# R2 Online Packet Documentation", "", "Generated automatically by `Scripts/generate_packet_docs.py`.", ""]
    for m in models:
        name = m['packet_enum']
        code = m['packet_code']
        header = f"## {code if code is not None else '?'} - {name}"
        lines.append(header)
        lines.append("")
        lines.append(f"**Direction:** {m['direction']}")
        lines.append("")
        lines.append(m['description'])
        lines.append("")
        if m['fields']:
            lines.append("| Field | Type |")
            lines.append("| --- | --- |")
            for field_type, field_name in m['fields']:
                lines.append(f"| {field_name} | {field_type} |")
            lines.append("")
        lines.append("")
    OUTPUT_PATH.parent.mkdir(exist_ok=True)
    OUTPUT_PATH.write_text('\n'.join(lines), encoding='utf-8')


if __name__ == '__main__':
    enum_map = parse_packet_enum()
    models = collect_models(enum_map)
    generate_docs(models)
    print(f"Documentation generated to {OUTPUT_PATH}")
