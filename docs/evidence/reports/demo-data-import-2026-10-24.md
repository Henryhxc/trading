# Demo Data Import

Date: 2026-10-24 (Asia/Singapore)

## Imported securities

| Symbol | Security | Daily bars | Range | Quote source/status |
| --- | --- | ---: | --- | --- |
| `09988` | Alibaba Group | 125 | 2026-04-24 to 2026-10-23 | `tencent` / `unofficial` |
| `01810` | Xiaomi Corporation | 125 | 2026-04-24 to 2026-10-23 | `tencent` / `unofficial` |

Source: Tencent public HK quote endpoint (`web.ifzq.gtimg.cn/appstock/app/fqkline/get`). This is an unofficial public source, not an HKEX licensed real-time feed. The last returned trading date was 2026-10-23; no future or synthetic bars were added.

## Demo watchlist

The existing local user `021021hxc@gmail.com` now has the following ordered watchlist:

1. `00700` Tencent
2. `09988` Alibaba Group
3. `01810` Xiaomi Corporation

The password is intentionally excluded from this evidence and from the repository. The database write was transactional and idempotent: repeated imports update the same stock/date keys and preserve the unique watchlist relationship.
