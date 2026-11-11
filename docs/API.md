## Popular catalog snapshot (10 November 2026)

`GET /api/games` now defaults to `popularity_order ASC NULLS LAST`, then release date/id descending. Optional JSON fields: `popularityOrder` (selection position), `popularitySource` (Most Played or Top Sellers TH), `popularityUpdatedAt` (import time). These describe a manually refreshed subset, not live player counts or global ranks. The frontend fetches once and derives search/genre/sort plus 24-item pagination locally.

The manual server-side maintenance command `go run ./cmd/import-popular` requires the dedicated GameMatch DATABASE_URL; it is not a public HTTP endpoint. It verifies 100 eligible entries before one DB transaction, rolls back the entire write on failure, and preserves descriptions/publication. Existing Admin Steam endpoint remains a separate 12-New-Releases action.

# API Contract
Base: REACT_APP_API_URL (เช่น http://localhost:8080). JSON body/response. Auth: `Authorization: Bearer TOKEN`.
Error ทุก handler: `{ "error": "ข้อความ" }` ยกเว้น ServeMux default 404/405 และ CORS rejection อาจเป็น text; FE มี fallback JSON handling.

| Method | Path | สิทธิ์ | ผลลัพธ์ |
|---|---|---|---|
| GET | /healthz | Public | status ok (503 หาก DB ไม่พร้อม) |
| POST | /api/auth/register | Public | id; body email,password,displayName |
| POST | /api/auth/login | Public | token,user; body email,password |
| POST | /api/auth/logout | Member | ยกเลิก JWT ของบัญชีทุก session |
| GET | /api/me | Member | id,email,displayName,role |
| PUT | /api/me | Member | body displayName |
| PUT | /api/me/password | Member | body currentPassword,newPassword; ยกเลิก session เดิม |
| GET | /api/games?q=keyword | Public | Game[] ที่เผยแพร่เท่านั้น |
| GET | /api/games/{id} | Public | Game |
| POST | /api/recommendations | Public | Game[] top5 พร้อม matchScore,reasons |
| GET/PUT | /api/me/preferences | Member | Preference object |
| GET | /api/me/favorites | Member | Game[] ของตนเองที่ยังเผยแพร่ |
| PUT/DELETE | /api/me/favorites/{id} | Member | ok; id คือ game id |
| GET | /api/admin/games | Admin | Game[] รวมเกมซ่อน |
| POST | /api/admin/games | Admin | Game body -> id |
| PUT | /api/admin/games/{id} | Admin | Game body -> id |
| DELETE | /api/admin/games/{id} | Admin | ok; ลบ favorites ที่อ้างอิงด้วย cascade |
| POST | /api/admin/sync | Admin | imported,note; RAWG first page 40 |
| POST | /api/admin/sync/steam | Admin | imported,note; Steam New Releases up to 12 eligible games, no RAWG key |

Member includes Admin. No endpoint accepts member_id from frontend for ownership.

Preference example:
```json
{"genre":"rpg","mode":"solo","style":"story"}
```
Allowed genre: rpg/action/strategy/simulation/adventure; mode: solo/coop/pvp; style: story/competitive/relax/challenge.
Mode is a **soft preference** worth 25, not a hard filter. Game with RPG+story but no confirmed solo tag can score 75. Generic multiplayer does not count as coop.

Game create example:
```json
{"title":"Example","description":"Example description","imageUrl":"","genres":["rpg"],"tags":["singleplayer","story-rich"],"platforms":["PC"],"isPublished":true}
```
Game response adds id, optional rawgId, matchScore, optional reasons. On PUT backend uses path id; role/owner never inferred from client body. Image URL if supplied must be HTTPS.

Steam responses also include `steamId`, ISO `releaseDate` (YYYY-MM-DD), and `syncedAt` (UTC timestamp). Catalog sorts release date descending, then id, with undated manual/RAWG records last. Steam sync reads the public Store New Releases feed for Thailand/English, inspects at most 20 candidates, and imports at most 12 released full games with valid Steam image/date and Store tags. Excludes DLC, coming-soon/future/ambiguous dates, missing tag pages and explicit sexual-content tags. Generic multiplayer never becomes co-op. HTML descriptions are converted to text; React never renders unsafe HTML. These public Store endpoints are not a documented/versioned Steamworks contract, so parsing/upstream failures return 502 and retain existing data. All writes use one transaction and unique steam_id; existing descriptions/publication are preserved. Sync is manual, not scheduled, and does not claim all new releases worldwide.

HTTP: 400 validation, 401 missing/expired/revoked token, 403 wrong role/origin, 404 missing record, 409 duplicate/concurrent operation, 500 DB failure, 502 upstream failure, 503 RAWG key missing/DB unavailable.

Login/Register share a limit of 30 attempts per TCP peer per minute. Limit responses are JSON 429 with `Retry-After` seconds. Forwarded headers cannot bypass the bucket; NAT/Render proxy users may share it. The limiter is in-memory for one process, resets on restart, and is not a distributed production limiter. Map capacity is bounded to 10,000 peers and expired entries are reclaimed.

RAWG sync rejects upstream error statuses, missing/null `results`, invalid/trailing JSON, bodies over 8 MiB, and responses with more than 40 entries. The SQL transaction is atomic; a database failure after an earlier insert rolls the entire import back. `RawgClient` is an optional Go test seam; production still calls the fixed RAWG HTTPS endpoint and uses a 20-second client timeout. The server DB pool is limited to five connections.

Scoring: Genre45 + Mode25 + Style30, sort score desc then title then id, exclude hidden and score0. PostgreSQL games -> Go rank -> JSON -> React. Not ML, not probability.

Frontend transport: `src/api/backend.js` ใช้ native fetch แนบ Bearer token เมื่อมี และส่ง Content-Type JSON เฉพาะเมื่อมี body; รองรับ AbortSignal, ลบ slash ท้าย base URL และแสดงข้อความเมื่อ network/non-JSON error เกิดขึ้น HTTP status ยังคงส่งต่อให้ AuthContext แยก 401 จาก server outage ได้ API หลังบ้านยังไม่เปลี่ยน contract ในรอบนี้

JSON ที่อ่านไม่สำเร็จจะ throw แม้ HTTP 200 เพื่อไม่ส่ง object ข้อผิดพลาดไปยังหน้าที่รอ Game[] และคง AbortError หากยกเลิกระหว่างอ่าน body หน้า Admin ส่งเฉพาะฟิลด์ที่แก้ไขได้และ trim arrays ตอน submit เสมอ ไม่ส่ง matchScore/reasons; หากเขียนสำเร็จแต่โหลดรายการใหม่ไม่ได้ ปุ่มโหลดรายการอีกครั้งเรียก GET โดยไม่ทำ POST/PUT/DELETE ซ้ำ

Frontend มีการตรวจ UTF-8 byte limits เพื่อแจ้งผู้ใช้ก่อนส่งฟอร์ม: displayName 100 bytes, title 250 bytes, password สูงสุด 72 bytes (พร้อม minLength 8 ของฟอร์มเดิม) Backend validation เป็นผู้บังคับข้อกำหนดจริงเสมอ

Session ฝั่ง frontend: 401 ของ protected API ล้าง token เฉพาะเมื่อยังตรงกับ token ตอนเริ่มคำขอ แล้วแจ้ง AuthContext ให้ล้าง user; public login 401 และ wrong-role 403 ไม่ล้าง session ที่มีอยู่ การเปลี่ยน token จากแท็บอื่นใช้ storage event เพื่อซ่อนข้อมูลเดิมและเรียก /api/me ใหม่ การตรวจบัญชีที่ได้ 5xx เก็บ token ไว้พร้อม retry ไม่ถือว่าเป็น revoked session ทั้งหมดเป็นพฤติกรรม UI เสริมจากการตรวจสิทธิ์ใน Go

Navigation หลังล็อกอิน: หน้าที่ขอเข้าสู่ระบบส่ง pathname ผ่าน Router state.from; สลับไปสมัครสมาชิกแล้วกลับ Login ยังคงหน้าปลายทางเดิม `auth/returnPath.js` ยอมรับเฉพาะ route ภายในที่มีจริง หากเป็น URL ภายนอกหรือ state ผิดรูปแบบจะกลับหน้าแรก ไม่เปลี่ยน API contract หรือเส้นทาง authorization ของ Go

Frontend routing: `/` เป็นหน้าแรก, `/discover` เป็นรายการเกม, `/about` อธิบายการใช้และแหล่งข้อมูล หน้า Home/Discover เรียก `GET /api/games` เพื่อโหลดรายการ published; ค้นหาชื่อและเลือกแนวในหน้า Discover เป็นการกรองจาก state ในเครื่อง Backend ยังรองรับ `?q=keyword` ตาม contract สำหรับผู้เรียกอื่นเหมือนเดิม ค่า return path ที่ไม่ผ่านการตรวจจะกลับ `/` (หน้าแรก)
