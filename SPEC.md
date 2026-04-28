# Sistema de Llamado de Mozos — Arquitectura DDD (Go + Next.js)

## 🧠 Contexto del dominio

Sistema para restaurantes donde:

- Cada mesa tiene un QR
- El cliente accede a una web app
- Puede llamar al mozo, pedir la cuenta o hacer consultas
- El restaurante recibe solicitudes en tiempo real
- Se registran métricas y feedback

---

## 🧱 Enfoque arquitectónico

Estilo: **DDD (Domain-Driven Design)** + **Clean Architecture**

```
/domain        → reglas de negocio puras
/application   → casos de uso
/infrastructure→ DB, WebSockets, APIs
/interfaces    → HTTP handlers / controllers
```

---

## 🧩 Dominio

### Restaurant

```go
type Restaurant struct {
    ID   string
    Name string
    Plan string
}
```

### Table

```go
type Table struct {
    ID           string
    Number       int
    RestaurantID string
    QRCode       string
}
```

### Request

```go
type RequestType string

const (
    CallWaiter RequestType = "CALL_WAITER"
    AskBill    RequestType = "ASK_BILL"
    AskHelp    RequestType = "ASK_HELP"
)

type RequestStatus string

const (
    Pending   RequestStatus = "PENDING"
    InProcess RequestStatus = "IN_PROCESS"
    Done      RequestStatus = "DONE"
)

type Request struct {
    ID        string
    TableID   string
    Type      RequestType
    Status    RequestStatus
    CreatedAt time.Time
}
```

### Feedback

```go
type Feedback struct {
    ID        string
    TableID   string
    Score     int
    CreatedAt time.Time
}
```

---

## ⚙️ Application Layer

### Casos de uso

- Crear solicitud
- Marcar como atendida
- Obtener solicitudes activas
- Registrar feedback

---

## 🏗 Infrastructure

- Base de datos: MySQL / PostgreSQL
- WebSockets (recomendado)
- Generación de QR dinámicos

---

## 🌐 Interfaces

### Cliente

- POST /requests
- GET /table/{id}/status
- POST /feedback

### Restaurante

- GET /requests/active
- PATCH /requests/{id}/complete

---

## 🖥 Frontend (Next.js)

### Cliente (vista de mesa — app del QR)

- Escanea QR → obtiene `qr_code` → llama a `POST /api/v1/session`
- Guarda `session_token` en memoria/sessionStorage
- Usa el token en `Authorization: Bearer <token>` para llamar a `POST /api/v1/requests`
- El token dura **30 minutos** — si expira (401), redirigir al QR

### Dashboard (panel del restaurante)

- Lista de mesas activas y sus solicitudes
- Conecta WebSocket por restaurante para recibir eventos en tiempo real
- Botón para marcar solicitudes como atendidas

---

## 🔌 API Reference

**Base URL:** `http://localhost:8080/api/v1`

### Autenticación de cliente (QR → JWT)

```
POST /session
Content-Type: application/json

{ "qr_code": "ABC12345XY" }
```

Respuesta `200`:

```json
{
  "session_token": "<jwt>",
  "table": { "number": 3 }
}
```

Errores:

- `400` — qr_code vacío
- `404` — QR no existe
- `403` — mesa inactiva

---

### Crear solicitud (cliente autenticado)

```
POST /requests
Authorization: Bearer <session_token>
Content-Type: application/json

{ "type": "CALL_WAITER" }
```

Tipos válidos: `CALL_WAITER` | `ASK_BILL` | `ASK_HELP`

Respuesta `201`:

```json
{
  "id": "uuid",
  "table_id": "uuid",
  "type": "CALL_WAITER",
  "status": "PENDING",
  "created_at": "2026-04-28T10:00:00Z"
}
```

Errores:

- `401` — sin token o expirado
- `400` — tipo inválido
- `429` — cooldown activo (header `Retry-After: <segundos>`)

---

### Obtener solicitudes activas (panel)

```
GET /restaurants/:restaurantId/requests/active
```

Respuesta `200`: array de `RequestResponse`

---

### Marcar solicitud como atendida (panel)

```
PATCH /requests/:requestId
Content-Type: application/json

{ "status": "DONE" }
```

Respuesta `200` o `422`

---

### Estado de solicitudes de una mesa

```
GET /tables/:tableId/status
```

---

### Feedback del cliente

```
POST /feedback
Content-Type: application/json

{
  "table_id": "uuid",
  "score": 5,
  "comment": "Excelente servicio"
}
```

Score: 1–5. Respuesta `201`.

---

### Regenerar QR de una mesa (admin)

```
POST /admin/tables/:id/regenerate-qr
```

Respuesta `200`: `{ "qr_code": "NUEVOCOD10" }`

---

## 📡 WebSocket (panel del restaurante)

### Conexión

```
ws://localhost:8080/api/v1/ws/:restaurantId?token=<session_token>
```

El `session_token` debe pertenecer a una mesa de ese restaurante (mismo `restaurant_id` en el JWT). Si el token es de otro restaurante → `403`.

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

- Go (Gin / Fiber)
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
