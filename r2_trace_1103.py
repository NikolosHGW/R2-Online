# r2_trace_1103.py
import frida
import sys
import time

TARGET = "R2ClientRU.exe"

JS = r"""
const ENABLE_PARSER_FALLBACK = false; // parser hook may destabilize client
const ENABLE_STATE_TRACE = true;      // dump candidate RC4 state near decrypt routine
const ENABLE_CALLCHAIN_TRACE = true;  // trace calls near 0x579Bxx pipeline

function rd8(p) {
    if (Memory && typeof Memory.readU8 === "function") return Memory.readU8(p);
    return p.readU8();
}
function rd16(p) {
    if (Memory && typeof Memory.readU16 === "function") return Memory.readU16(p);
    return p.readU16();
}
function rd32(p) {
    if (Memory && typeof Memory.readU32 === "function") return Memory.readU32(p);
    return p.readU32();
}
function rdPtr(p) {
    if (Memory && typeof Memory.readPointer === "function") return Memory.readPointer(p);
    return p.readPointer();
}
function rdCStr(p) {
    if (Memory && typeof Memory.readCString === "function") return Memory.readCString(p);
    return p.readCString();
}

function u16le(p) { return rd16(p); }
function toLowerSafe(x) { return (x || "").toString().toLowerCase(); }

let gLastWelcome = null;
let gStateTraceHits = 0;
let gDidWelcomeScan = false;
let gLastPktStart = null;
let gLastPktEnd = null;

function tryGetExport(moduleName, exportName) {
    try {
        return Module.getExportByName(moduleName, exportName);
    } catch (e) {
        try {
            return Module.findExportByName(moduleName, exportName);
        } catch (e2) {
            try {
                return Module.findExportByName(null, exportName);
            } catch (e3) {
                return null;
            }
        }
    }
}

function dumpWelcomeFromPtr(pkt, tag, ctx) {
    try {
        const size = u16le(pkt);
        if (size < 6 || size > 4096) return;

        const enc = rd8(pkt.add(2));
        const seq = rd8(pkt.add(3));
        const op = u16le(pkt.add(4));
        if (op !== 0x044f) return;

        let bt = [];
        try {
            if (Thread && Thread.backtrace && DebugSymbol && DebugSymbol.fromAddress) {
                bt = Thread.backtrace(ctx, Backtracer.FUZZY)
                    .map(function (a) { return DebugSymbol.fromAddress(a).toString(); });
            }
        } catch (e) {
            // keep bt empty
        }

        send({
            type: "hit",
            tag,
            pkt: pkt.toString(),
            pkt_size: size,
            enc,
            seq,
            op,
            backtrace: bt
        });

        try {
            if (typeof hexdump === "function") {
                send({
                    type: "hex",
                    data: hexdump(pkt, {
                        offset: 0,
                        length: Math.min(size, 224),
                        header: true,
                        ansi: false
                    })
                });
            }
        } catch (e) {
            // ignore hexdump issues
        }

        const payload = pkt.add(6);
        gLastPktStart = ptr(pkt);
        gLastPktEnd = ptr(pkt).add(size);
        try {
            gLastWelcome = [];
            for (let i = 0; i < size - 6; i++) gLastWelcome.push(rd8(payload.add(i)));
        } catch (e) {
            gLastWelcome = null;
        }
        if (!gDidWelcomeScan && gLastWelcome) {
            gDidWelcomeScan = true;
            try { scanProcessForWelcomeChunks(pkt, size); } catch (e) {}
        }
        function dump18(off, name) {
            if (size < (6 + off + 18)) return;
            let out = [];
            for (let i = 0; i < 18; i++) {
                const b = rd8(payload.add(off + i));
                out.push(("0" + b.toString(16)).slice(-2));
            }
            send({ type: "key", name, offset: off, bytes: out.join(" ") });
        }

        dump18(96,  "mKey8_guess_96");
        dump18(153, "mKey8_guess_153");
        dump18(155, "mKey8_guess_155");
    } catch (e) {
        send({ type: "warn", where: "dumpWelcomeFromPtr", err: e.toString() });
    }
}

function bytesPattern(arr, off, n) {
    const out = [];
    for (let i = 0; i < n; i++) out.push(("0" + arr[off + i].toString(16)).slice(-2));
    return out.join(" ");
}

function scanProcessForWelcomeChunks(pkt, size) {
    if (!gLastWelcome || gLastWelcome.length < 174) return;
    const chunks = [
        { name: "wel96", off: 96 },
        { name: "wel153", off: 153 },
        { name: "wel155", off: 155 },
    ];
    let ranges = [];
    try {
        // Frida compatibility: some builds accept only string form.
        if (typeof Process.enumerateRangesSync === "function") {
            ranges = Process.enumerateRangesSync({ protection: "rw-", coalesce: true });
            if (!ranges || ranges.length === 0) ranges = Process.enumerateRangesSync("rw-");
        } else if (typeof Process.enumerateRanges === "function") {
            ranges = Process.enumerateRanges("rw-");
        } else {
            throw new Error("enumerateRanges API not available");
        }
    } catch (e) {
        send({ type: "warn", where: "scanProcessForWelcomeChunks.enumerateRanges", err: e.toString() });
        return;
    }
    const pktStart = ptr(pkt);
    const pktEnd = pktStart.add(size);
    send({ type: "state", where: "scan", ptr: pktStart.toString(), note: "rw ranges=" + ranges.length, dump: "" });
    for (let ci = 0; ci < chunks.length; ci++) {
        const c = chunks[ci];
        const pat = bytesPattern(gLastWelcome, c.off, 18);
        let total = 0;
        for (let ri = 0; ri < ranges.length; ri++) {
            const r = ranges[ri];
            // Skip the recv packet buffer range to reduce obvious duplicates.
            const rStart = ptr(r.base);
            const rEnd = rStart.add(r.size);
            if (pktStart.compare(rEnd) < 0 && pktEnd.compare(rStart) > 0) continue;
            let hits = [];
            try { hits = Memory.scanSync(r.base, r.size, pat); } catch (e) { continue; }
            for (let hi = 0; hi < hits.length; hi++) {
                total++;
                if (total <= 20) {
                    send({
                        type: "state",
                        where: "scan-hit",
                        ptr: hits[hi].address.toString(),
                        note: c.name + " off=" + c.off,
                        dump: bytesHex(hits[hi].address, 24)
                    });
                }
            }
        }
        send({ type: "state", where: "scan-summary", ptr: "0x0", note: c.name, dump: "hits=" + total });
    }
}

function targetedProbeAround(ptrBase, tag) {
    if (!gLastWelcome || !looksPtr(ptrBase)) return;
    const base = ptr(ptrBase);
    const chunks = [
        { name: "wel96", off: 96 },
        { name: "wel153", off: 153 },
        { name: "wel155", off: 155 },
    ];
    for (let ci = 0; ci < chunks.length; ci++) {
        const c = chunks[ci];
        const pat = bytesPattern(gLastWelcome, c.off, 18);
        // local window is fast and low-risk
        for (let w = -0x4000; w <= 0x4000; w += 0x2000) {
            const start = base.add(w);
            try {
                const hits = Memory.scanSync(start, 0x2000, pat);
                for (let hi = 0; hi < hits.length; hi++) {
                    if (gLastPktStart && gLastPktEnd) {
                        if (hits[hi].address.compare(gLastPktStart) >= 0 && hits[hi].address.compare(gLastPktEnd) < 0) {
                            continue; // skip obvious packet buffer matches
                        }
                    }
                    send({
                        type: "state",
                        where: "target-hit",
                        ptr: hits[hi].address.toString(),
                        note: tag + " " + c.name + " win=" + w,
                        dump: bytesHex(hits[hi].address, 24)
                    });
                }
            } catch (e) {}
        }
    }
    // Also inspect pointer fields inside base object.
    for (let off = 0; off <= 0x200; off += 4) {
        let q = null;
        try { q = rdPtr(base.add(off)); } catch (e) { continue; }
        if (!looksPtr(q)) continue;
        try {
            send({
                type: "state",
                where: "obj-ptr",
                ptr: q.toString(),
                note: tag + " off=0x" + off.toString(16),
                dump: bytesHex(q, 32)
            });
        } catch (e) {}
        traceCandidateState(tag + ".obj+" + off.toString(16), q);
    }
}

function bytesHex(p, n) {
    const out = [];
    for (let i = 0; i < n; i++) out.push(("0" + rd8(p.add(i)).toString(16)).slice(-2));
    return out.join(" ");
}

function isPermutation256(p) {
    try {
        const seen = {};
        for (let i = 0; i < 256; i++) {
            const b = rd8(p.add(i));
            if (seen[b]) return false;
            seen[b] = true;
        }
        return true;
    } catch (e) {
        return false;
    }
}

function isPermutation256DwordLowByte(p) {
    try {
        const seen = {};
        for (let i = 0; i < 256; i++) {
            const v = rd32(p.add(i * 4));
            const b = v & 0xff;
            if (seen[b]) return false;
            seen[b] = true;
        }
        return true;
    } catch (e) {
        return false;
    }
}

function traceCandidateState(tag, basePtr) {
    if (!basePtr || basePtr.isNull()) return;
    const b = ptr(basePtr);
    try {
        if (isPermutation256(b)) {
            send({ type: "state", where: tag, ptr: b.toString(), note: "perm@+0", dump: bytesHex(b, 64) });
        }
    } catch (e) {}
    try {
        if (isPermutation256(b.add(8))) {
            send({ type: "state", where: tag, ptr: b.add(8).toString(), note: "perm@+8", dump: bytesHex(b.add(8), 64) });
        }
    } catch (e) {}
    // CRc4A keeps S as 256 dwords (low byte is actual permutation).
    try {
        if (isPermutation256DwordLowByte(b)) {
            send({ type: "state", where: tag, ptr: b.toString(), note: "perm32@+0(lowbyte)", dump: bytesHex(b, 64) });
        }
    } catch (e) {}
    try {
        if (isPermutation256DwordLowByte(b.add(8))) {
            send({ type: "state", where: tag, ptr: b.add(8).toString(), note: "perm32@+8(lowbyte)", dump: bytesHex(b.add(8), 64) });
        }
    } catch (e) {}

    // One-level pointer chasing: many builds keep RC4 state behind object pointers.
    for (let off = 0; off <= 0x80; off += 4) {
        let q = null;
        try { q = rdPtr(b.add(off)); } catch (e) { continue; }
        if (!looksPtr(q)) continue;
        try {
            if (isPermutation256(q)) {
                send({ type: "state", where: tag, ptr: q.toString(), note: "perm@ptr+" + off.toString(16), dump: bytesHex(q, 64) });
            }
        } catch (e) {}
        try {
            if (isPermutation256(q.add(8))) {
                send({ type: "state", where: tag, ptr: q.add(8).toString(), note: "perm@ptr+" + off.toString(16) + "+8", dump: bytesHex(q.add(8), 64) });
            }
        } catch (e) {}
        try {
            if (isPermutation256DwordLowByte(q)) {
                send({ type: "state", where: tag, ptr: q.toString(), note: "perm32@ptr+" + off.toString(16), dump: bytesHex(q, 64) });
            }
        } catch (e) {}
        try {
            if (isPermutation256DwordLowByte(q.add(8))) {
                send({ type: "state", where: tag, ptr: q.add(8).toString(), note: "perm32@ptr+" + off.toString(16) + "+8", dump: bytesHex(q.add(8), 64) });
            }
        } catch (e) {}
    }
}

function maybeTraceDecryptState(ctx) {
    if (!ENABLE_STATE_TRACE) return;
    if (gStateTraceHits >= 8) return;
    gStateTraceHits++;

    const ecx = ptr(ctx.ecx);
    const esp = ptr(ctx.esp);
    const slots = [];
    try { slots.push(ptr(rd32(esp.add(0x4)))); } catch (e) {}
    try { slots.push(ptr(rd32(esp.add(0x8)))); } catch (e) {}
    try { slots.push(ptr(rd32(esp.add(0xC)))); } catch (e) {}
    try { slots.push(ptr(rd32(esp.add(0x10)))); } catch (e) {}
    try { slots.push(ptr(rd32(esp.add(0x14)))); } catch (e) {}

    send({
        type: "state",
        where: "enter decrypt candidate",
        ptr: ecx.toString(),
        note: "ecx/stack",
        dump: "ecx=" + ecx.toString() + " args=" + slots.map(function (x) { return x.toString(); }).join(",")
    });

    traceCandidateState("ecx", ecx);
    for (let i = 0; i < slots.length; i++) traceCandidateState("arg" + (i + 1), slots[i]);

    // If welcome was captured, scan small windows around candidates for known 18-byte chunks.
    if (gLastWelcome && gLastWelcome.length >= 174) {
        const chunks = [
            { name: "wel+96", off: 96 },
            { name: "wel+153", off: 153 },
            { name: "wel+155", off: 155 },
        ];
        function hasChunk(p, off) {
            try {
                for (let i = 0; i < 18; i++) {
                    if (rd8(p.add(i)) !== gLastWelcome[off + i]) return false;
                }
                return true;
            } catch (e) { return false; }
        }
        const allPtrs = [ecx].concat(slots);
        for (let pi = 0; pi < allPtrs.length; pi++) {
            const bp = allPtrs[pi];
            if (!bp || bp.isNull()) continue;
            for (let delta = -0x40; delta <= 0x40; delta += 4) {
                const p = bp.add(delta);
                for (let ci = 0; ci < chunks.length; ci++) {
                    const c = chunks[ci];
                    if (hasChunk(p, c.off)) {
                        send({
                            type: "state",
                            where: "chunk-hit",
                            ptr: p.toString(),
                            note: c.name + " near ptr#" + pi + " delta=" + delta,
                            dump: bytesHex(p, 24)
                        });
                    }
                }
            }
        }
    }
}

function looksPtr(p) {
    if (!p || p.isNull()) return false;
    try { return p.compare(ptr("0x10000")) >= 0; } catch (e) { return false; }
}

function traceRegs(tag, ctx) {
    try {
        const regs = {
            eax: ptr(ctx.eax), ebx: ptr(ctx.ebx), ecx: ptr(ctx.ecx), edx: ptr(ctx.edx),
            esi: ptr(ctx.esi), edi: ptr(ctx.edi), ebp: ptr(ctx.ebp), esp: ptr(ctx.esp),
        };
        send({
            type: "state",
            where: tag,
            ptr: regs.ecx.toString(),
            note: "regs",
            dump: Object.keys(regs).map(function (k) { return k + "=" + regs[k].toString(); }).join(" ")
        });
        const candidates = [regs.eax, regs.ebx, regs.ecx, regs.edx, regs.esi, regs.edi];
        for (let i = 0; i < candidates.length; i++) {
            if (looksPtr(candidates[i])) traceCandidateState(tag + ".r" + i, candidates[i]);
        }
    } catch (e) {}
}

function scanFor1103(baseBuf, n, tag, ctx) {
    if (!baseBuf || n < 6) return;
    for (let off = 0; off <= n - 6; off++) {
        const p = baseBuf.add(off);
        const size = u16le(p);
        if (size < 6) continue;
        if (off + size > n) continue;
        const op = u16le(p.add(4));
        if (op === 0x044f) {
            dumpWelcomeFromPtr(p, tag + "+scan", ctx);
        }
    }
}

const attached = { recv: false, wsaRecv: false, recvfrom: false, parser: false };
let warnedNoHooks = false;
let warnedImports = false;
const hookedAddrs = {};

function fnKind(name) {
    const n = toLowerSafe(name);
    if (n === "recv" || n === "_recv" || n.indexOf("recv@16") !== -1 || n.indexOf("_recv@16") !== -1) return "recv";
    if (n === "wsarecv" || n === "_wsarecv" || n.indexOf("wsarecv@32") !== -1 || n.indexOf("_wsarecv@32") !== -1) return "wsaRecv";
    if (n === "recvfrom" || n === "_recvfrom" || n.indexOf("recvfrom@24") !== -1 || n.indexOf("_recvfrom@24") !== -1) return "recvfrom";
    return null;
}

function attachAt(kind, addr, tag) {
    if (!addr || addr.isNull()) return false;
    const k = addr.toString();
    if (hookedAddrs[k]) return true;

    try {
        if (kind === "recv") {
            Interceptor.attach(addr, {
                onEnter(args) { this.buf = args[1]; },
                onLeave(retval) {
                    const n = retval.toInt32();
                    if (n > 0) scanFor1103(this.buf, n, tag + "!recv", this.context);
                }
            });
            attached.recv = true;
        } else if (kind === "wsaRecv") {
            Interceptor.attach(addr, {
                onEnter(args) {
                    this.lpBuffers = args[1];
                    this.dwBufferCount = args[2].toInt32();
                    this.lpNumberOfBytesRecvd = args[3];
                },
                onLeave(retval) {
                    if (retval.toInt32() !== 0) return;
                    if (this.dwBufferCount <= 0) return;
                    let total = 0;
                try { total = rd32(this.lpNumberOfBytesRecvd); } catch (e) { return; }
                    if (total <= 0) return;
                    const offBuf = (Process.pointerSize === 8) ? 8 : 4;
                const len = rd32(this.lpBuffers);
                const buf = rdPtr(this.lpBuffers.add(offBuf));
                    const n = Math.min(total, len);
                    if (n > 0) scanFor1103(buf, n, tag + "!WSARecv", this.context);
                }
            });
            attached.wsaRecv = true;
        } else if (kind === "recvfrom") {
            Interceptor.attach(addr, {
                onEnter(args) { this.buf = args[1]; },
                onLeave(retval) {
                    const n = retval.toInt32();
                    if (n > 0) scanFor1103(this.buf, n, tag + "!recvfrom", this.context);
                }
            });
            attached.recvfrom = true;
        } else {
            return false;
        }

        hookedAddrs[k] = true;
        send({ type: "hook", name: tag + "@" + kind, addr: k });
        return true;
    } catch (e) {
        return false;
    }
}

function attachRecv(moduleName) {
    if (attached.recv) return;
    const p = tryGetExport(moduleName, "recv");
    if (!p) return;

    if (!attachAt("recv", p, moduleName || "null")) {
        send({ type: "warn", where: "attachRecv(" + moduleName + ")", err: "attach failed" });
    }
}

function attachWSARecv(moduleName) {
    if (attached.wsaRecv) return;
    const p = tryGetExport(moduleName, "WSARecv");
    if (!p) return;

    if (!attachAt("wsaRecv", p, moduleName || "null")) {
        send({ type: "warn", where: "attachWSARecv(" + moduleName + ")", err: "attach failed" });
    }
}

function attachRecvFrom(moduleName) {
    if (attached.recvfrom) return;
    const p = tryGetExport(moduleName, "recvfrom");
    if (!p) return;

    if (!attachAt("recvfrom", p, moduleName || "null")) {
        send({ type: "warn", where: "attachRecvFrom(" + moduleName + ")", err: "attach failed" });
    }
}

function attachByResolver() {
    let r = null;
    try { r = new ApiResolver("module"); } catch (e) { return; }

    function firstNetExport(name) {
        const q = "exports:*!"+name;
        let list = [];
        try { list = r.enumerateMatchesSync(q); } catch (e) { return null; }
        for (let i = 0; i < list.length; i++) {
            const n = (list[i].name || "").toLowerCase();
            if (n.indexOf("ws2_32.dll!") !== -1 || n.indexOf("wsock32.dll!") !== -1) {
                return list[i].address;
            }
        }
        return null;
    }

    if (!attached.recv) {
        const p = firstNetExport("recv");
        if (p) {
            try {
                Interceptor.attach(p, {
                    onEnter(args) { this.buf = args[1]; },
                    onLeave(retval) {
                        const n = retval.toInt32();
                        if (n > 0) scanFor1103(this.buf, n, "resolver!recv", this.context);
                    }
                });
                attached.recv = true;
                send({ type: "hook", name: "resolver!recv", addr: p.toString() });
            } catch (e) {}
        }
    }

    if (!attached.wsaRecv) {
        const p = firstNetExport("WSARecv");
        if (p) {
            try {
                Interceptor.attach(p, {
                    onEnter(args) {
                        this.lpBuffers = args[1];
                        this.dwBufferCount = args[2].toInt32();
                        this.lpNumberOfBytesRecvd = args[3];
                    },
                    onLeave(retval) {
                        if (retval.toInt32() !== 0) return;
                        if (this.dwBufferCount <= 0) return;
                        let total = 0;
                        try { total = rd32(this.lpNumberOfBytesRecvd); } catch (e) { return; }
                        if (total <= 0) return;
                        const offBuf = (Process.pointerSize === 8) ? 8 : 4;
                        const len = rd32(this.lpBuffers);
                        const buf = rdPtr(this.lpBuffers.add(offBuf));
                        const n = Math.min(total, len);
                        if (n > 0) scanFor1103(buf, n, "resolver!WSARecv", this.context);
                    }
                });
                attached.wsaRecv = true;
                send({ type: "hook", name: "resolver!WSARecv", addr: p.toString() });
            } catch (e) {}
        }
    }

    if (!attached.recvfrom) {
        const p = firstNetExport("recvfrom");
        if (p) {
            try {
                Interceptor.attach(p, {
                    onEnter(args) { this.buf = args[1]; },
                    onLeave(retval) {
                        const n = retval.toInt32();
                        if (n > 0) scanFor1103(this.buf, n, "resolver!recvfrom", this.context);
                    }
                });
                attached.recvfrom = true;
                send({ type: "hook", name: "resolver!recvfrom", addr: p.toString() });
            } catch (e) {}
        }
    }
}

function attachByExportScan() {
    const mods = ["ws2_32.dll", "mswsock.dll", "wsock32.dll"];
    for (let mi = 0; mi < mods.length; mi++) {
        let m = null;
        try { m = Process.getModuleByName(mods[mi]); } catch (e) { continue; }
        if (!m) continue;
        let ex = [];
        try { ex = m.enumerateExports(); } catch (e) { continue; }
        for (let i = 0; i < ex.length; i++) {
            const kind = fnKind(ex[i].name || "");
            if (!kind) continue;
            attachAt(kind, ex[i].address, "exportscan:" + mods[mi]);
        }
    }
}

function nameLooksLike(n, base) {
    if (!n) return false;
    const s = n.toLowerCase();
    return s === base || s === ("_" + base) || s.indexOf(base + "@") !== -1 || s.indexOf("_" + base + "@") !== -1;
}

function attachFromMainImports() {
    let main = null;
    try { main = Process.enumerateModules()[0]; } catch (e) { return; }
    if (!main) return;

    let imports = [];
    try { imports = main.enumerateImports(); } catch (e) { return; }

    if (!warnedImports) {
        warnedImports = true;
        try {
            const recvLike = imports
                .filter(function (imp) {
                    const n = (imp.name || "").toLowerCase();
                    return n.indexOf("recv") !== -1;
                })
                .map(function (imp) {
                    return (imp.module || "?") + "!" + (imp.name || "?");
                });
            send({ type: "warn", where: "imports", err: "recv-like imports: " + recvLike.join(", ") });
        } catch (e) {}
    }

    function attachImport(baseName, key, onEnter, onLeave, tag) {
        if (attached[key]) return;

        for (let i = 0; i < imports.length; i++) {
            const imp = imports[i];
            if (!nameLooksLike(imp.name || "", baseName)) continue;

            let target = null;
            try { target = rdPtr(imp.address); } catch (e) { continue; }
            if (!target || target.isNull()) continue;

            try {
                Interceptor.attach(target, { onEnter: onEnter, onLeave: onLeave });
                attached[key] = true;
                send({ type: "hook", name: tag + " (" + (imp.module || "?") + "!" + imp.name + ")", addr: target.toString() });
                return;
            } catch (e) {
                // try next import
            }
        }
    }

    attachImport(
        "recv",
        "recv",
        function (args) { this.buf = args[1]; },
        function (retval) {
            const n = retval.toInt32();
            if (n > 0) scanFor1103(this.buf, n, "imports!recv", this.context);
        },
        "imports!recv"
    );

    attachImport(
        "wsarecv",
        "wsaRecv",
        function (args) {
            this.lpBuffers = args[1];
            this.dwBufferCount = args[2].toInt32();
            this.lpNumberOfBytesRecvd = args[3];
        },
        function (retval) {
            if (retval.toInt32() !== 0) return;
            if (this.dwBufferCount <= 0) return;
            let total = 0;
            try { total = rd32(this.lpNumberOfBytesRecvd); } catch (e) { return; }
            if (total <= 0) return;
            const offBuf = (Process.pointerSize === 8) ? 8 : 4;
            const len = rd32(this.lpBuffers);
            const buf = rdPtr(this.lpBuffers.add(offBuf));
            const n = Math.min(total, len);
            if (n > 0) scanFor1103(buf, n, "imports!WSARecv", this.context);
        },
        "imports!WSARecv"
    );

    attachImport(
        "recvfrom",
        "recvfrom",
        function (args) { this.buf = args[1]; },
        function (retval) {
            const n = retval.toInt32();
            if (n > 0) scanFor1103(this.buf, n, "imports!recvfrom", this.context);
        },
        "imports!recvfrom"
    );
}

function attachViaGetProcAddress() {
    if (hookedAddrs["getproc"]) return;

    function hookOne(mod, name) {
        const p = tryGetExport(mod, name);
        if (!p) return false;
        try {
            Interceptor.attach(p, {
                onEnter(args) {
                    this.namePtr = args[1];
                    this.reqName = null;
                    try {
                        if (!this.namePtr.isNull()) this.reqName = rdCStr(this.namePtr);
                    } catch (e) {}
                },
                onLeave(retval) {
                    if (!this.reqName) return;
                    const kind = fnKind(this.reqName);
                    if (!kind) return;
                    attachAt(kind, retval, "GetProcAddress");
                }
            });
            hookedAddrs["getproc"] = true;
            send({ type: "hook", name: mod + "!" + name, addr: p.toString() });
            return true;
        } catch (e) {
            return false;
        }
    }

    if (hookOne("kernel32.dll", "GetProcAddress")) return;
    hookOne("kernelbase.dll", "GetProcAddress");
}

function attachParserFallback() {
    if (attached.parser) return;

    let main = null;
    try {
        main = Process.enumerateModules()[0];
    } catch (e) {
        send({ type: "warn", where: "attachParserFallback.enumerateModules", err: e.toString() });
        return;
    }
    if (!main) {
        send({ type: "warn", where: "attachParserFallback", err: "main module not found" });
        return;
    }

    // x32dbg: VA 0x00579B71, assumed base 0x00400000 => RVA 0x179B71
    const addr = main.base.add(0x179B71);
    const k = addr.toString();
    try {
        Interceptor.attach(addr, {
            onEnter(args) {
                try {
                    const esi = ptr(this.context.esi);
                    const op = rd16(esi.add(0x2A));
                    if (op !== 0x044f) return;
                    const pkt = esi.add(0x2A).sub(4);
                    dumpWelcomeFromPtr(pkt, "parser@" + k, this.context);
                } catch (e) {
                    // not a packet parser context
                }
            }
        });
        attached.parser = true;
        send({ type: "hook", name: "parser-fallback", addr: k });
    } catch (e) {
        send({ type: "warn", where: "attachParserFallback", err: e.toString() });
    }
}

function attachStateTraceHook() {
    if (!ENABLE_STATE_TRACE) return;
    if (hookedAddrs["statehook"]) return;
    let main = null;
    try { main = Process.enumerateModules()[0]; } catch (e) { return; }
    if (!main) return;

    // Candidates from observed call-chain/backtrace in this client build.
    const rvas = [0x135E0, 0x179B44, 0x17A058, 0x17A27D];
    let ok = 0;
    for (let i = 0; i < rvas.length; i++) {
        const addr = main.base.add(rvas[i]);
        const name = "statehook(main+0x" + rvas[i].toString(16) + ")";
        try {
            Interceptor.attach(addr, {
                onEnter(args) {
                    try {
                        maybeTraceDecryptState(this.context);
                        // If this frame sees welcome packet layout through ESI, dump it.
                        const esi = ptr(this.context.esi);
                        if (!esi.isNull()) {
                            const op = rd16(esi.add(0x2A));
                            if (op === 0x044f) {
                                const pkt = esi.add(0x2A).sub(4);
                                dumpWelcomeFromPtr(pkt, "state@" + name, this.context);
                            }
                        }
                    } catch (e) {}
                }
            });
            ok++;
            send({ type: "hook", name: name, addr: addr.toString() });
        } catch (e) {
            // ignore single bad candidate
        }
    }
    if (ok > 0) {
        hookedAddrs["statehook"] = true;
    } else {
        send({ type: "warn", where: "attachStateTraceHook", err: "no candidate attached" });
    }
}

function attachCallChainHooks() {
    if (!ENABLE_CALLCHAIN_TRACE) return;
    let main = null;
    try { main = Process.enumerateModules()[0]; } catch (e) { return; }
    if (!main) return;

    const chain = [
        { rva: 0x1799A0, name: "cc_5799A0" },
        { rva: 0x121420, name: "cc_521420" },
        { rva: 0x121330, name: "cc_521330" },
        { rva: 0x0135E0, name: "cc_4135E0" },
    ];
    for (let i = 0; i < chain.length; i++) {
        const c = chain[i];
        const addr = main.base.add(c.rva);
        const hk = "cc_" + c.rva.toString(16);
        if (hookedAddrs[hk]) continue;
        try {
            Interceptor.attach(addr, {
                onEnter(args) {
                    try {
                        // For cc_5799A0, EAX often points to packet opcode (buf+4).
                        // Filter noise from intro/UI threads and trace only packet-like calls.
                        if (c.name === "cc_5799A0") {
                            let op = -1;
                            try { op = rd16(ptr(this.context.eax)); } catch (e) {}
                            if (op !== 0x044f && op !== 0x0c1c) return;
                            send({ type: "state", where: "packet-op", ptr: ptr(this.context.eax).toString(), note: c.name, dump: "op=0x" + op.toString(16) });
                            try {
                                targetedProbeAround(ptr(this.context.ecx), "cc_5799A0.ecx");
                                targetedProbeAround(ptr(this.context.esi), "cc_5799A0.esi");
                                targetedProbeAround(ptr(this.context.edx), "cc_5799A0.edx");
                            } catch (e) {}
                        }
                        traceRegs("callchain:" + c.name, this.context);
                        maybeTraceDecryptState(this.context);
                    } catch (e) {}
                }
            });
            hookedAddrs[hk] = true;
            send({ type: "hook", name: "callchain(" + c.name + ")", addr: addr.toString() });
        } catch (e) {}
    }
}

function tryAttachAll() {
    [null, "ws2_32.dll", "WS2_32.DLL", "wsock32.dll", "WSOCK32.DLL"].forEach(function(m) {
        attachRecv(m);
        attachWSARecv(m);
        attachRecvFrom(m);
    });
    attachByResolver();
    attachByExportScan();
    attachFromMainImports();
    attachViaGetProcAddress();
    attachStateTraceHook();
    attachCallChainHooks();
    if (ENABLE_PARSER_FALLBACK) {
        attachParserFallback();
    }

    if (!attached.recv && !attached.wsaRecv && !attached.recvfrom && !warnedNoHooks) {
        warnedNoHooks = true;
        try {
            const mods = Process.enumerateModules()
                .map(function (m) { return m.name.toLowerCase(); })
                .filter(function (n) { return n.indexOf("ws") !== -1 || n.indexOf("sock") !== -1; });
            send({ type: "warn", where: "tryAttachAll", err: "no network hooks yet; modules=" + mods.join(",") });
        } catch (e) {
            send({ type: "warn", where: "tryAttachAll", err: "no network hooks yet" });
        }
    }
}

tryAttachAll();
setInterval(function () { tryAttachAll(); }, 500);

send({
    type: "ready",
    arch: Process.arch,
    pointerSize: Process.pointerSize,
    mainModule: Process.enumerateModules()[0].name
});
"""


def on_message(message, data):
    if message["type"] == "send":
        p = message["payload"]
        t = p.get("type")
        if t == "ready":
            print(
                f"[*] script loaded, waiting for hooks... "
                f"(arch={p.get('arch')} ptr={p.get('pointerSize')} main={p.get('mainModule')})"
            )
        elif t == "hook":
            print(f"[*] hooked {p['name']} @ {p['addr']}")
        elif t == "hit":
            print("\n=== HIT 1103 ===")
            print(
                f"src={p['tag']} pkt={p['pkt']} size={p['pkt_size']} "
                f"enc={p['enc']} seq={p['seq']} op=0x{p['op']:04X}"
            )
            print("backtrace:")
            for x in p["backtrace"]:
                print("  ", x)
        elif t == "hex":
            print(p["data"])
        elif t == "key":
            print(f"{p['name']} (payload+{p['offset']}): {p['bytes']}")
        elif t == "warn":
            print(f"[warn] {p['where']}: {p['err']}")
        elif t == "state":
            print(f"[state] {p.get('where')} ptr={p.get('ptr')} {p.get('note')}")
            d = p.get("dump")
            if d:
                print("        ", d)
        else:
            print(p)
    else:
        print(message)


def pick_pid(dev):
    # usage: py r2_trace_1103.py <PID>
    if len(sys.argv) > 1 and sys.argv[1].isdigit():
        return int(sys.argv[1])

    procs = [p for p in dev.enumerate_processes() if "r2clientru.exe" in p.name.lower()]
    if not procs:
        return None

    print("[*] candidates:")
    for p in procs:
        print(f"    pid={p.pid} name={p.name}")

    return max(procs, key=lambda p: p.pid).pid


def main():
    dev = frida.get_local_device()
    pid = pick_pid(dev)

    if pid is None:
        print(f"[!] process not found: {TARGET}")
        sys.exit(1)

    print(f"[*] attach pid={pid}")
    session = dev.attach(pid)
    script = session.create_script(JS)
    script.on("message", on_message)
    script.load()

    try:
        while True:
            time.sleep(0.2)
    except KeyboardInterrupt:
        pass


if __name__ == "__main__":
    main()
