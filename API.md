# Waiter API Reference

**Base:** `http://localhost:8080/api/v1`  
**Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## Públicas

### `GET /health`

Health check.

```json
// 200
{ "status": "ok" }
```

### `POST /session` — Iniciar sesión con QR

```json
// Request
{ "qr_code": "ABC12345XY" }

// 200
{ "session_token": "<jwt>", "table": { "number": 3 } }
```

Errores: `400` · `404`

### `GET /tables/:tableId/status` — Estado de mesa

Retorna solicitudes activas de la mesa. No requiere auth.

### `POST /feedback` — Registrar feedback (público)

```json
// Request
{ "table_id": "uuid", "score": 5, "comment": "Excelente" }

// 201
{ "id": "uuid", "table_id": "uuid", "score": 5, "comment": "Excelente", "created_at": "..." }
```

Errores: `400` · `422`

---

## Cliente (requieren `Authorization: Bearer <session_token>`)

### `POST /requests` — Crear solicitud

```json
// Request
{ "type": "CALL_WAITER" }

// 201
{ "id": "uuid", "table_id": "uuid", "type": "CALL_WAITER", "status": "PENDING", "created_at": "..." }
```

Tipos: `CALL_WAITER` | `ASK_BILL` | `ASK_HELP`  
Errores: `401` · `400` · `429` (cooldown 15s, header `Retry-After`)

---

## Admin (requieren `Authorization: Bearer <admin_token>`)

### `POST /admin/login` — Login

```json
// Request
{ "username": "admin", "password": "secret" }

// 200
{ "token": "<jwt>" }
```

Errores: `400` · `401`

### `GET /restaurants` — Listar restaurantes _(superadmin)_

### `POST /restaurants` — Crear restaurante _(superadmin)_

```json
// Request
{ "name": "La Trattoria", "plan": "pro" }

// 201
{ "id": "uuid", "name": "La Trattoria", "plan": "pro", "qr_banner_text": "Escaneá y llamá al mozo", "qr_footer_items": ["Llamar al mozo", "Pedir la cuenta", "Dejar reseña"] }
```

Campos QR opcionales: `qr_banner_text` (max 100 chars), `qr_footer_items` (array 1-5 items, cada uno max 50 chars).  
Errores: `400` · `422`

### `GET /restaurants/:restaurantId` — Obtener restaurante _(superadmin, owner)_

### `POST /restaurants/:restaurantId/tables` — Crear mesa _(superadmin, owner)_

```json
// Request
{ "number": 5 }

// 201
{ "id": "uuid", "number": 5, "restaurant_id": "uuid", "qr_code": "XXXX" }
```

Errores: `409` duplicado · `422`

### `GET /restaurants/:restaurantId/tables` — Listar mesas _(superadmin, owner)_

### `POST /restaurants/:restaurantId/tables/:id/inactivate` — Inactivar mesa _(superadmin, owner)_

```json
// 200
{ "table_id": "uuid", "is_active": false }
```

### `POST /restaurants/:restaurantId/tables/:id/activate` — Activar mesa _(superadmin, owner)_

```json
// 200
{ "table_id": "uuid", "is_active": true }
```

### `POST /admin/tables/:id/regenerate-qr` — Regenerar QR _(superadmin, owner)_

```json
// 200
{ "qr_code": "NUEVOCOD" }
```

### `GET /restaurants/:restaurantId/requests/active` — Solicitudes activas _(superadmin, owner, employee)_

### `PATCH /requests/:requestId` — Actualizar estado _(cualquier admin)_

```json
// Request
{ "status": "DONE" }

// 204
```

### `POST /admin/users` — Crear usuario _(superadmin, owner)_

```json
// Request
{ "username": "juan", "password": "secret", "role": "employee", "restaurant_id": "uuid" }

// 201
{ "id": "uuid", "username": "juan", "role": "employee", "restaurant_id": "uuid", "created_at": "..." }
```

Errores: `400` · `403` · `422`

### `GET /admin/users` — Listar usuarios _(superadmin, owner)_

Superadmin ve todos; owner ve solo los de su restaurante.

### `PATCH /admin/users/:id` — Modificar usuario _(superadmin, owner)_

Campos opcionales (solo enviar lo que se quiere cambiar).

```json
// Request (uno o ambos campos)
{ "username": "nuevo-nombre" }

// 200
{ "id": "uuid", "username": "nuevo-nombre", "role": "employee", "restaurant_id": "uuid", "created_at": "..." }
```

Errores: `400` · `403` · `404`

### `POST /admin/users/:id/reset-password` — Blanquear contraseña _(superadmin, owner)_

Genera y retorna una nueva contraseña aleatoria.

```json
// 200
{ "new_password": "a1b2c3d4e5f6g7h8" }
```

Errores: `403` · `404`

### `DELETE /admin/users/:id` — Eliminar usuario _(superadmin, owner)_

Superadmin elimina cualquiera; owner solo employees de su restaurante.  
`204` sin body. Errores: `403` · `404`

---

## WebSocket

### `GET /ws/:restaurantId?token=<jwt>`

Acepta admin token o session token cuyo `restaurant_id` coincida con `:restaurantId`.

**Eventos recibidos:**

```json
// new_request
{ "event": "new_request", "request": { "id": "uuid", "table_id": "uuid", "type": "CALL_WAITER", "status": "PENDING", "created_at": "..." } }

// request_completed
{ "event": "request_completed", "request_id": "uuid" }
```
