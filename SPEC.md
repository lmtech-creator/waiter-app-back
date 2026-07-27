# Sistema de Llamado de Mozos

Sistema para restaurantes donde cada mesa tiene un código QR. El cliente escanea el QR, obtiene un JWT de sesión y puede llamar al mozo, pedir la cuenta o dejar feedback. El panel de administración recibe solicitudes en tiempo real vía WebSocket.

---

## Arquitectura

Estilo: **Clean Architecture / DDD**

```
domain/entity          → structs de dominio (sin lógica)
domain/repository      → interfaces de repositorio
application/usecase    → lógica de negocio
interfaces/http        → Gin handlers, DTOs, middleware
infrastructure/        → GORM/PostgreSQL, JWT, WebSocket hub
cmd/server/main.go     → wiring completo
```

**Stack:** Go 1.25 · Gin · GORM/PostgreSQL · JWT HS256 · Zap · Swagger

---

## Dominio

### Entidades

| Entidad      | Campos clave                                                          |
| ------------ | --------------------------------------------------------------------- |
| `Restaurant` | `id`, `name`, `plan`, `qr_banner_text`, `qr_footer_items`              |
| `Table`      | `id`, `number`, `restaurant_id`, `qr_code`, `is_active`               |
| `Request`    | `id`, `table_id`, `type`, `status`, `created_at`                      |
| `Feedback`   | `id`, `table_id`, `score` (1–5), `comment`, `created_at`              |
| `AdminUser`  | `id`, `username`, `password_hash`, `role`, `restaurant_id` (nullable) |

### Tipos

```
RequestType:   CALL_WAITER | ASK_BILL | ASK_HELP
RequestStatus: PENDING | IN_PROCESS | DONE
AdminRole:     superadmin | owner | employee
```

---

## Auth — dos tokens independientes

|                       | Cliente (sesión QR)                         | Admin                                 |
| --------------------- | ------------------------------------------- | ------------------------------------- |
| Endpoint de obtención | `POST /session`                             | `POST /admin/login`                   |
| Secret env var        | `SESSION_SECRET`                            | `ADMIN_SECRET`                        |
| Duración              | 30 min                                      | 24 h                                  |
| Header                | `Authorization: Bearer <session_token>`     | `Authorization: Bearer <admin_token>` |
| Claims                | `table_id`, `restaurant_id`, `table_number` | `admin_id`, `restaurant_id`, `role`   |

---

## RBAC — roles de administrador

| Rol          | `restaurant_id` | Acceso                             |
| ------------ | --------------- | ---------------------------------- |
| `superadmin` | `null`          | Todo el sistema                    |
| `owner`      | requerido       | Solo su restaurante                |
| `employee`   | requerido       | Rutas operativas de su restaurante |

### Tabla de rutas por rol

| Método   | Ruta                                         | Roles permitidos                                        |
| -------- | -------------------------------------------- | ------------------------------------------------------- |
| `POST`   | `/admin/login`                               | público                                                 |
| `GET`    | `/restaurants`                               | superadmin                                              |
| `POST`   | `/restaurants`                               | superadmin                                              |
| `GET`    | `/restaurants/:restaurantId`                 | superadmin, owner (propio)                              |
| `POST`   | `/restaurants/:restaurantId/tables`          | superadmin, owner (propio)                              |
| `GET`    | `/restaurants/:restaurantId/tables`          | superadmin, owner (propio)                              |
| `POST`   | `/admin/tables/:id/regenerate-qr`            | superadmin, owner (propio)                              |
| `POST`   | `/restaurants/:restaurantId/tables/:id/inactivate` | superadmin, owner (propio)                              |
| `POST`   | `/restaurants/:restaurantId/tables/:id/activate`   | superadmin, owner (propio)                              |
| `GET`    | `/restaurants/:restaurantId/requests/active` | superadmin, owner, employee (propio)                    |
| `PATCH`  | `/requests/:requestId`                       | superadmin, owner, employee                             |
| `POST`   | `/admin/users`                               | superadmin, owner (solo employees propios)              |
| `GET`    | `/admin/users`                               | superadmin (todos), owner (propios)                     |
| `DELETE` | `/admin/users/:id`                           | superadmin (cualquiera), owner (solo employees propios) |

---

## API Reference

**Base URL:** `http://localhost:8080/api/v1`  
**Swagger UI:** `http://localhost:8080/swagger/index.html`

---

### Rutas públicas

#### `GET /health` — Health check

```json
// Response 200
{ "status": "ok" }
```

---

#### `POST /session` — Iniciar sesión con QR

```json
// Request
{ "qr_code": "ABC12345XY" }

// Response 200
{ "session_token": "<jwt>", "table": { "number": 3 } }
```

Errores: `400` qr_code vacío · `404` QR no existe

---

#### `GET /tables/:tableId/status` — Estado de mesa

Retorna solicitudes activas de la mesa. No requiere auth.

---

#### `POST /feedback` — Registrar feedback

```json
// Request
{ "table_id": "uuid", "score": 5, "comment": "Excelente" }

// Response 201
{ "id": "uuid", "table_id": "uuid", "score": 5, "comment": "Excelente", "created_at": "..." }
```

Errores: `400` · `422`

---

### Rutas de cliente (requieren `session_token`)

#### `POST /requests` — Crear solicitud

```json
// Request
{ "type": "CALL_WAITER" }

// Response 201
{ "id": "uuid", "table_id": "uuid", "type": "CALL_WAITER", "status": "PENDING", "created_at": "..." }
```

Tipos válidos: `CALL_WAITER` | `ASK_BILL` | `ASK_HELP`  
Errores: `401` · `400` tipo inválido · `429` cooldown (header `Retry-After: <seg>`)

---

### Rutas de admin (requieren `admin_token`)

#### `POST /admin/login`

```json
// Request
{ "username": "admin", "password": "secret" }

// Response 200
{ "token": "<jwt>" }
```

Errores: `400` · `401`

---

#### `GET /restaurants` — Listar restaurantes _(superadmin)_

Respuesta `200`: array de `RestaurantResponse`

---

#### `POST /restaurants` — Crear restaurante _(superadmin)_

```json
// Request
{ "name": "La Trattoria", "plan": "pro" }

// Response 201
{ "id": "uuid", "name": "La Trattoria", "plan": "pro" }
```

---

#### `GET /restaurants/:restaurantId` — Obtener restaurante _(superadmin, owner)_

---

#### `POST /restaurants/:restaurantId/tables` — Crear mesa _(superadmin, owner)_

```json
// Request
{ "number": 5 }

// Response 201
{ "id": "uuid", "number": 5, "restaurant_id": "uuid", "qr_code": "XXXX" }
```

Errores: `409` número duplicado · `422`

---

#### `GET /restaurants/:restaurantId/tables` — Listar mesas _(superadmin, owner)_

---

#### `POST /admin/tables/:id/regenerate-qr` — Regenerar QR _(superadmin, owner)_

```json
// Response 200
{ "qr_code": "NUEVOCOD" }
```

---

#### `GET /restaurants/:restaurantId/requests/active` — Solicitudes activas _(superadmin, owner, employee)_

---

#### `PATCH /requests/:requestId` — Actualizar estado _(cualquier admin)_

```json
// Request
{ "status": "DONE" }

// Response 204
```

---

#### `POST /admin/users` — Crear usuario admin _(superadmin, owner)_

```json
// Request
{ "username": "juan", "password": "secret", "role": "employee", "restaurant_id": "uuid" }

// Response 201
{ "id": "uuid", "username": "juan", "role": "employee", "restaurant_id": "uuid", "created_at": "..." }
```

Reglas: superadmin puede crear cualquier rol; owner solo puede crear `employee` de su propio restaurante.  
Errores: `400` · `403` · `422` username duplicado

---

#### `GET /admin/users` — Listar usuarios admin _(superadmin, owner)_

Superadmin ve todos; owner ve solo los de su restaurante.

---

#### `DELETE /admin/users/:id` — Eliminar usuario admin _(superadmin, owner)_

Superadmin puede eliminar cualquiera; owner solo puede eliminar employees de su restaurante.  
Respuesta `204`. Errores: `403` · `404`

---

### WebSocket

#### `GET /ws/:restaurantId`

Acepta `?token=<jwt>` — intenta admin token primero, luego session token. El `restaurant_id` del claims debe coincidir con `:restaurantId`.  
Emite eventos en tiempo real cuando se crean o actualizan solicitudes.

---

## Seed inicial

Al arrancar, si no existe ningún admin, el sistema crea automáticamente un `superadmin` y loguea las credenciales:

```
INFO  superadmin created — change this password immediately
      username: admin
      password: <generado>
```

---

## Variables de entorno

| Variable          | Descripción                                | Default     |
| ----------------- | ------------------------------------------ | ----------- |
| `DATABASE_URL`    | PostgreSQL DSN                             | —           |
| `SESSION_SECRET`  | Secret HMAC para tokens de cliente         | —           |
| `ADMIN_SECRET`    | Secret HMAC para tokens de admin           | —           |
| `ALLOWED_ORIGINS`           | Orígenes permitidos CORS (comma-separated)   | acepta todo |
| `PORT`                      | Puerto del servidor                          | `8080`      |
| `SEED_ADMIN_RESTAURANT_ID` | UUID del restaurante para el admin inicial   | —           |

---

## 📡 WebSocket (panel del restaurante)

### Conexión

```
ws://localhost:8080/api/v1/ws/:restaurantId?token=<session_token>
```

Acepta JWT de admin (firmado con `ADMIN_SECRET`) o de sesión de cliente (firmado con `SESSION_SECRET`). El `restaurant_id` del claims debe coincidir con `:restaurantId`. Si el token es de otro restaurante → `403`.

### Eventos recibidos

**`new_request`** — cuando una mesa llama:

```json
{
  "event": "new_request",
  "request": {
    "id": "uuid",
    "table_id": "uuid",
    "type": "CALL_WAITER",
    "status": "PENDING",
    "created_at": "2026-04-28T10:00:00Z"
  }
}
```

**`request_completed`** — cuando se atiende una solicitud:

```json
{
  "event": "request_completed",
  "request_id": "uuid"
}
```

### Ejemplo de cliente JS

```js
// 1. Obtener un session_token válido para este restaurante
//    (cualquier mesa activa sirve — el dashboard puede tener el suyo propio)
const token = localStorage.getItem("session_token");

const ws = new WebSocket(
  `ws://localhost:8080/api/v1/ws/${restaurantId}?token=${token}`,
);

ws.onopen = () => console.log("Panel conectado");

ws.onmessage = ({ data }) => {
  const msg = JSON.parse(data);
  if (msg.event === "new_request") {
    // Agregar a la lista de pendientes, reproducir sonido, etc.
    addPending(msg.request);
  }
  if (msg.event === "request_completed") {
    removePending(msg.request_id);
  }
};

ws.onclose = () => {
  // Reconectar con backoff exponencial
  setTimeout(connectWS, 3000);
};
```

### Seguridad WS

| Mecanismo                 | Comportamiento                                                 |
| ------------------------- | -------------------------------------------------------------- |
| `?token` faltante         | 401 — conexión rechazada                                       |
| Token inválido/expirado   | 401 — conexión rechazada                                       |
| Token de otro restaurante | 403 — conexión rechazada                                       |
| Origin no permitido       | 403 — handshake rechazado (configurable via `ALLOWED_ORIGINS`) |

---

Cliente → Backend → Evento → Restaurante → Atención

---

## 🚀 Stack

Backend:

- Go (Gin)
- WebSockets

Frontend:

- Next.js
- Tailwind

Infra:

- Docker
- Nginx

---

## 🧠 Resumen

Sistema basado en:

- Eventos simples
- Estado controlado
- Tiempo real
- Dominio claro
