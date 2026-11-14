# Primary course reference

The user's highest-priority course reference for this project is:
https://github.com/nonthawat2548/week6-lab/tree/main/moviehub

Use its actual source code and README as the first reference for backend structure, dependencies, handlers, middleware, SQL, authentication, and configuration. Other course references below are supplementary. If the reference code and README disagree, explain the discrepancy and preserve working GameMatch behavior while implementing the user's requested scope. Do not claim the entire project complies without reviewing and testing it.

The following development rules were supplied by the user.
# GAMEMATCH — BACKEND STRICT DEVELOPMENT RULES

## 1. ข้อกำหนดสูงสุด (Non-Negotiable Rules)

GameMatch เป็น Term Project วิชา Backend และ Frontend Programming ระดับมหาวิทยาลัย ไม่ใช่ระบบ Enterprise

การพัฒนา Backend ต้องยึดเนื้อหา วิธีการเขียนโค้ด และเทคโนโลยีที่อาจารย์สอนเป็นหลัก โดยมีเป้าหมายให้ระบบทำงานได้จริง ครบตามเกณฑ์ประเมิน และนักศึกษาสามารถอธิบายโค้ดทุกส่วนได้

**หากวิธีการที่ทันสมัยหรือมีประสิทธิภาพมากกว่า ขัดกับวิธีการที่เรียน ให้เลือกวิธีการที่เรียนก่อนเสมอ**

## 2. แหล่งอ้างอิงหลัก

ให้ใช้ Repository ต่อไปนี้เป็นมาตรฐานในการพัฒนา Backend

- https://github.com/nonthawat2548/week6-lab
- https://github.com/nonthawat2548/week6-lab-fork
- Backend ของโปรเจกต์ GameMatch ปัจจุบัน
- เอกสารประกอบการสอนวิชา Backend Programming ที่ผู้ใช้ให้ไว้

ก่อนเขียนโค้ดใหม่ ต้องศึกษาโครงสร้างและรูปแบบการเขียนจากแหล่งอ้างอิงจริง

ห้ามสมมติเองว่าอาจารย์สอนเทคโนโลยีหรือวิธีการใด หากยังไม่ได้ตรวจสอบหลักฐานจากโค้ดหรือเอกสารประกอบการเรียน

## 3. Technology Stack

เทคโนโลยีหลักที่ต้องรักษาไว้:

- ภาษา Go (Golang)
- PostgreSQL
- REST API
- JSON
- HTTP Request และ Response
- Supabase สำหรับ PostgreSQL Database
- Render สำหรับ Backend Deployment
- GitHub สำหรับ Version Control

สำหรับ Router, Database Driver, Authentication Library และ Dependency อื่น ๆ ให้ยึดตามที่ปรากฏในโปรเจกต์อ้างอิงและโปรเจกต์ปัจจุบันเท่านั้น

### ห้ามใช้โดยพลการ

- Node.js / Express.js
- Python / Django / FastAPI
- Java / Spring Boot
- Gin, Fiber หรือ Echo หากไม่ได้ใช้ในบทเรียน
- GORM หรือ ORM อื่นที่ไม่ได้เรียน
- Prisma
- GraphQL
- gRPC
- Redis
- MongoDB
- Firebase Authentication
- Supabase Authentication เพื่อแทนระบบ Authentication เดิม
- WebSocket
- Microservices
- Message Queue
- AI API หรือ Machine Learning
- Library และ Dependency ที่ไม่ได้รับอนุมัติ

ห้ามเปลี่ยนภาษา Framework หรือ Database เพียงเพราะ AI คุ้นเคยกับเทคโนโลยีอื่นมากกว่า

## 4. รูปแบบการเขียนโค้ด

ต้องรักษารูปแบบการเขียน Backend ให้สอดคล้องกับที่เรียน เช่น:

- Struct
- Function
- Method
- Package
- Error Handling
- JSON Marshal / Unmarshal หรือ Encoder / Decoder ตามที่เรียน
- HTTP Handler
- Middleware ตามที่เรียน
- การเชื่อมต่อ PostgreSQL
- SQL Query
- CRUD Operation

### ข้อห้าม

- ห้ามออกแบบ Architecture ใหม่โดยไม่จำเป็น
- ห้ามเพิ่ม Design Pattern ที่ไม่ได้เรียน
- ห้ามสร้าง Abstraction หลายชั้นโดยไม่มีเหตุผล
- ห้ามแยก Service, Repository, Interface หรือ Layer เพิ่มเติมเพียงเพื่อให้ดูเป็นมืออาชีพ
- ห้ามเขียนโค้ดซับซ้อนเกินความสามารถในการอธิบายของนักศึกษา
- ห้ามใช้เทคนิคขั้นสูงที่ไม่ได้ปรากฏในบทเรียนโดยพลการ

หากสามารถแก้ปัญหาด้วย Function ธรรมดาได้ ให้ใช้ Function ธรรมดา

โค้ดต้องอ่านง่าย เข้าใจง่าย ดูแลรักษาได้ และสอดคล้องกับรูปแบบเดิม

## 5. Database Rules

PostgreSQL ต้องเป็นแหล่งข้อมูลหลักของระบบ

- ใช้ Primary Key และ Foreign Key อย่างถูกต้อง
- รักษาความสัมพันธ์ระหว่างตาราง
- ใช้ Data Type ให้เหมาะสม
- กำหนด Constraint ตามความจำเป็น
- ตรวจสอบความถูกต้องของข้อมูลก่อนบันทึก
- ใช้ Parameterized Query เพื่อป้องกัน SQL Injection
- จัดการ Database Error อย่างเหมาะสม
- รักษาความสอดคล้องระหว่าง Database Schema, Go Struct และ JSON Response
- ใช้ Transaction เมื่อมีการเปลี่ยนแปลงข้อมูลหลายส่วนที่ต้องสำเร็จพร้อมกัน

### ห้ามเด็ดขาด

- ห้าม Hardcode ข้อมูลเกมแทน Database
- ห้ามใช้ Array หรือ Slice เป็นฐานข้อมูลถาวร
- ห้ามใช้ Mock Data ในระบบ Production
- ห้ามต่อ SQL Query จาก Input ของผู้ใช้โดยตรง
- ห้าม Reset Database
- ห้าม Drop Table โดยไม่ได้รับอนุญาต
- ห้ามลบข้อมูลเดิมเพื่อแก้ปัญหา
- ห้ามเปลี่ยน Database Schema โดยไม่ตรวจสอบผลกระทบ

หากจำเป็นต้องเปลี่ยน Schema ต้องอธิบายเหตุผล ผลกระทบ และวิธีรักษาข้อมูลเดิมก่อนดำเนินการ

## 6. REST API Rules

Backend ต้องใช้ REST API ตามรูปแบบที่เรียน

HTTP Method ต้องตรงกับวัตถุประสงค์:

- GET — อ่านข้อมูล
- POST — สร้างข้อมูล
- PUT / PATCH — แก้ไขข้อมูลตามรูปแบบเดิมของโปรเจกต์
- DELETE — ลบข้อมูล

HTTP Status Code ต้องเหมาะสมกับผลลัพธ์ เช่น:

- 200 OK
- 201 Created
- 204 No Content
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 409 Conflict
- 500 Internal Server Error

ทุก Endpoint ต้อง:

1. ตรวจสอบข้อมูลที่ได้รับ
2. ตรวจสอบ Authentication หากจำเป็น
3. ตรวจสอบ Authorization หากจำเป็น
4. ดำเนินการตาม Business Logic
5. ติดต่อ Database ตามความเหมาะสม
6. ส่ง Response กลับอย่างถูกต้อง

ห้ามเปลี่ยน Endpoint, Request Body, JSON Field หรือ Response Structure เดิมโดยไม่ตรวจสอบ Frontend ที่เรียกใช้งาน

ห้ามตอบ HTTP 200 เมื่อการดำเนินการล้มเหลวจริง

## 7. Authentication และ Authorization

ต้องรักษาระบบ Authentication ตามที่เรียนและที่มีอยู่ในโปรเจกต์เดิม

หากระบบใช้ JWT และ bcrypt อยู่แล้ว ให้รักษาวิธีการดังกล่าวไว้

ระบบต้องแบ่งสิทธิ์เป็น 3 ระดับ:

### Guest

- อ่านข้อมูลเกมสาธารณะ
- ค้นหาเกม
- ดูรายละเอียดเกม
- สมัครสมาชิก
- เข้าสู่ระบบ

### Member

- ใช้งาน API สาธารณะ
- จัดการข้อมูลส่วนตัวของตนเอง
- จัดการความชอบของตนเอง
- จัดการรายการเกมโปรดของตนเอง
- รับคำแนะนำเกมตามความชอบ

### Admin

- จัดการข้อมูลเกม
- เพิ่ม แก้ไข และลบข้อมูลตามขอบเขตระบบ
- จัดการข้อมูลส่วนอื่นตามสิทธิ์ที่ได้รับอนุมัติ

### Security Rules

- ห้ามเก็บรหัสผ่านแบบ Plain Text
- ห้ามส่ง Password หรือ Password Hash กลับไปยัง Frontend
- ห้ามเชื่อถือ Role ที่ผู้ใช้ส่งมาใน Request
- ต้องตรวจสอบ Token จาก Backend
- ต้องตรวจสอบสิทธิ์ Admin จาก Backend
- ต้องตรวจสอบ Ownership ของข้อมูล
- Member ต้องไม่สามารถแก้ไขข้อมูลของ Member คนอื่น
- ผู้สมัครสมาชิกต้องไม่สามารถกำหนด Role ของตนเองเป็น Admin
- ห้ามใช้ Frontend Route Protection แทน Backend Authorization
- ห้ามข้าม Authentication เพื่อให้การทดสอบผ่าน

## 8. Game Recommendation Logic

GameMatch เป็นระบบแนะนำเกมตามความสนใจของผู้ใช้

Recommendation System ต้องใช้ Business Logic ที่เขียนด้วย Go และ SQL ตามขอบเขตที่เรียน

ข้อมูลที่สามารถใช้พิจารณา ได้แก่:

- Genre
- Platform
- Game Mode
- Play Style
- User Preferences
- ข้อมูลเกมอื่นที่มีอยู่จริงใน Database

สามารถใช้ Rule-based Recommendation หรือการคำนวณคะแนนความตรงกันตามเงื่อนไขได้

### ข้อห้าม

- ห้ามใช้ Machine Learning
- ห้ามใช้ AI API
- ห้ามใช้ Vector Database
- ห้ามใช้ Embedding
- ห้ามใช้ Recommendation Library ภายนอก
- ห้ามสุ่มเกมแล้วอ้างว่าเป็นระบบแนะนำ
- ห้ามสร้างคะแนนความตรงกันปลอม
- ห้าม Hardcode ผลลัพธ์เฉพาะผู้ใช้

ผลลัพธ์ต้องมาจากข้อมูลจริง และสามารถอธิบายได้ว่าทำไมเกมนั้นจึงได้รับการแนะนำ

ห้ามเปลี่ยนหลักการ Recommendation เดิมโดยไม่ได้รับอนุญาต

## 9. Error Handling

ต้องจัดการ Error ทุกจุดที่มีโอกาสเกิดความผิดพลาดตามความเหมาะสม เช่น:

- Request Body ไม่ถูกต้อง
- JSON Decode ไม่สำเร็จ
- Token ไม่ถูกต้อง
- Token หมดอายุ
- ไม่พบข้อมูล
- ข้อมูลซ้ำ
- Database Connection ล้มเหลว
- SQL Query ล้มเหลว
- ผู้ใช้ไม่มีสิทธิ์เข้าถึง

ห้ามละเลย Error โดยใช้ `_` ทิ้งในจุดสำคัญ

ห้ามใช้ panic เพื่อจัดการ Error ทั่วไปของ API

ห้ามส่งรายละเอียด Database Error, Secret Key หรือ Stack Trace ที่มีข้อมูลภายในให้ Client

Error Response ต้องมีรูปแบบที่ Frontend สามารถนำไปใช้งานได้จริง

## 10. Environment Variables และ Security Configuration

ข้อมูลสำคัญต้องจัดเก็บผ่าน Environment Variables ตามโครงสร้างเดิมของระบบ

ตัวอย่าง:

- DATABASE_URL
- JWT_SECRET
- PORT
- Configuration อื่นที่ระบบใช้งานจริง

### ห้ามเด็ดขาด

- ห้าม Hardcode Database Password
- ห้าม Hardcode JWT Secret
- ห้าม Commit ไฟล์ `.env` ที่มีข้อมูลจริง
- ห้ามเปิดเผย Supabase Database Credentials
- ห้ามใส่ Secret ใน Source Code
- ห้ามบันทึกรหัสผ่านหรือ Token ลง Log โดยไม่จำเป็น
- ห้ามส่ง Database Credentials ไปยัง Frontend

สามารถใช้ `.env.example` เพื่อแสดงตัวอย่างการตั้งค่าได้ แต่ต้องไม่มี Secret จริง

## 11. CORS และ Frontend Integration

Backend ต้องสามารถทำงานร่วมกับ Frontend ที่ Deploy อยู่จริง

- ตรวจสอบ Allowed Origin
- ตรวจสอบ HTTP Methods
- ตรวจสอบ Allowed Headers
- รองรับ Authorization Header ตามระบบเดิม
- จัดการ OPTIONS Preflight ตามความเหมาะสม
- รักษา API Contract เดิม

ห้ามแก้ปัญหา CORS ด้วยการปิดระบบรักษาความปลอดภัยทั้งหมด

หากแก้ไข Backend API ต้องตรวจสอบว่า Frontend ยังเรียกใช้งานได้ตามปกติ

ห้ามเปลี่ยน Frontend เพื่อปกปิดข้อผิดพลาดของ Backend

## 12. Testing Rules

หลังแก้ไข Backend ต้องตรวจสอบและทดสอบส่วนที่เกี่ยวข้องจริง

อย่างน้อยให้พิจารณา:

- `go build ./...`
- `go test ./...`
- `go vet ./...`
- `gofmt` สำหรับไฟล์ Go ที่แก้ไข
- Database Connection
- Registration
- Login
- JWT Validation
- Role Authorization
- CRUD Operations
- Recommendation
- Error Response

การทดสอบ API ต้องครอบคลุมทั้งกรณีสำเร็จและกรณีล้มเหลว

ตัวอย่างกรณีที่ต้องตรวจสอบ:

- Login ด้วยรหัสผ่านผิด
- เรียก API โดยไม่มี Token
- Member เรียก Admin API
- Member แก้ไขข้อมูลของผู้อื่น
- ค้นหาเกมที่ไม่มีอยู่
- ส่ง Request Body ไม่ถูกต้อง
- เพิ่มข้อมูลที่ขัดกับ Database Constraint

ห้ามลบ Test เพื่อให้ผลทดสอบผ่าน

ห้ามรายงานว่า Test ผ่านหากยังไม่ได้รันจริง

## 13. Deployment Rules

โครงสร้าง Backend Deployment ต้องรักษาไว้ดังนี้:

- Source Code: GitHub
- Backend Hosting: Render
- Database: Supabase PostgreSQL

ต้องตรวจสอบการทำงานของ:

- Build Command
- Start Command
- Environment Variables
- Database Connection
- PORT
- CORS
- Production API URL
- Deployment Logs

ห้ามเปลี่ยนผู้ให้บริการ Deployment โดยพลการ

ห้าม Hardcode localhost สำหรับ Production

ห้ามสร้าง Database หรือ Backend Service ใหม่เพื่อหลีกเลี่ยงการแก้ปัญหาของระบบเดิม

## 14. ห้ามแก้ไขเกินขอบเขตงาน

ก่อนแก้ไขต้องอ่านและตรวจสอบ:

- `go.mod`
- โครงสร้าง Backend
- Routes
- Handlers
- Models หรือ Struct
- Middleware
- Database Schema
- API Contract
- Environment Configuration
- Tests ที่เกี่ยวข้อง

ห้ามสร้างไฟล์หรือ Dependency เพิ่มโดยไม่มีเหตุผล

ห้าม Refactor ทั้ง Backend หากผู้ใช้ขอแก้เพียงฟังก์ชันเดียว

ห้ามลบฟังก์ชันเดิมที่ยังใช้งานได้

ห้ามเปลี่ยนชื่อ Route, Function, Struct หรือ Database Field โดยไม่ตรวจสอบการอ้างอิงทั้งหมด

ห้ามแก้ไข Frontend หากได้รับคำสั่งให้ทำเฉพาะ Backend เว้นแต่จำเป็นและได้รับอนุญาต

## 15. ขั้นตอนการทำงานของ AI

AI ต้องดำเนินงานตามลำดับดังนี้:

1. อ่านโค้ดปัจจุบัน
2. ตรวจสอบโปรเจกต์อ้างอิงจากรายวิชา
3. วิเคราะห์ต้นเหตุของปัญหา
4. ตรวจสอบว่าการแก้ไขอยู่ภายในขอบเขตที่เรียน
5. ระบุไฟล์ที่ได้รับผลกระทบ
6. ลงมือแก้ไขจริง
7. ตรวจสอบความถูกต้องของโค้ด
8. Build และ Test
9. ตรวจสอบผลกระทบกับระบบเดิม
10. รายงานสิ่งที่เปลี่ยนแปลง

หากสามารถแก้ไขและทดสอบด้วยเครื่องมือที่มีอยู่ได้ ให้ดำเนินการเอง ไม่ใช่เพียงส่งคำสั่งให้ผู้ใช้ทำทุกขั้นตอน

ห้ามส่งโค้ดตัวอย่างแทนการแก้ไขจริง หากผู้ใช้สั่งให้แก้ไขโปรเจกต์โดยตรง

ห้ามอ้างว่าแก้ไขสำเร็จ หากยังไม่ได้แก้ไขจริง

## 16. รายงานหลังแก้ไข

เมื่อทำงานเสร็จ ต้องรายงาน:

- ปัญหาที่พบ
- สาเหตุของปัญหา
- วิธีแก้ไข
- ไฟล์ที่แก้ไข
- Endpoint ที่ได้รับผลกระทบ
- Database Schema ที่เปลี่ยนแปลง (ถ้ามี)
- ผลการ Build
- ผลการ Test
- สิ่งที่ยังไม่ได้ตรวจสอบ
- ข้อจำกัดหรือความเสี่ยงที่เหลืออยู่

ห้ามปกปิด Error หรือรายงานความสำเร็จเกินจริง

## 17. เกณฑ์ตัดสินใจขั้นสุดท้าย

เมื่อมีหลายวิธีในการแก้ปัญหา ให้เรียงลำดับความสำคัญดังนี้:

1. ตรงกับสิ่งที่อาจารย์สอน
2. ครบตามเกณฑ์ Term Project
3. ระบบทำงานได้ถูกต้อง
4. มีความปลอดภัย
5. ไม่ทำลายระบบเดิม
6. นักศึกษาสามารถอธิบายโค้ดได้
7. โค้ดอ่านง่ายและดูแลรักษาได้
8. ประสิทธิภาพเหมาะสม

หากข้อกำหนดใดขัดแย้งกัน ให้แจ้งผู้ใช้และอธิบายเหตุผล ห้ามตัดสินใจเปลี่ยนเทคโนโลยีหรือขอบเขตโปรเจกต์เอง

**หลักการสูงสุด: ใช้สิ่งที่เรียนให้เกิดประโยชน์สูงสุด ไม่ใช่นำสิ่งที่ไม่ได้เรียนมาแก้ปัญหาแทน**

**Do not overengineer. Do not introduce unapproved technologies. Follow the course code conventions. Preserve existing functionality. Test before claiming success.**