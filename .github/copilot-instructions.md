## Proyecto

**Sistema de llamado de mozos para restaurantes** — Go backend con REST + WebSocket.
Módulo: `github.com/waiter/back` · Go 1.25 · Gin · GORM/PostgreSQL · JWT (HS256) · Zap

---

## Arquitectura

Capas en orden de dependencia (las superiores dependen de las inferiores):

| Capa          | Paquete                      | Responsabilidad                                                                           |
| ------------- | ---------------------------- | ----------------------------------------------------------------------------------------- |
| Entidades     | `domain/entity`              | Structs puros sin lógica (Restaurant, Table, Request, Feedback, AdminUser)                |
| Repositorios  | `domain/repository`          | Interfaces únicamente — sin implementaciones                                              |
| Casos de uso  | `application/usecase`        | Lógica de negocio; recibe interfaces, no implementaciones concretas                       |
| Handlers HTTP | `interfaces/http`            | Gin handlers + DTOs en `interfaces/http/dto`                                              |
| Middleware    | `interfaces/http/middleware` | `AuthMiddleware` (cliente), `AdminAuthMiddleware` (admin), `RateLimitMiddleware`          |
| Persistencia  | `infrastructure/persistence` | GORM + PostgreSQL; AutoMigrate en `database.go`                                           |
| Auth          | `infrastructure/auth`        | `SignSession`/`VerifySession` (cliente) + `SignAdminSession`/`VerifyAdminSession` (admin) |
| WebSocket     | `infrastructure/websocket`   | Hub de habitaciones por `restaurantId`                                                    |
| Mocks         | `mocks/mocks.go`             | Un único archivo con todos los mocks — actualizar siempre que cambie una interfaz         |
| Entrada       | `cmd/server/main.go`         | Wiring completo: DB → repos → usecases → handlers → `SetupRouter`                         |

### Regla de dependencias

Las capas superiores **no importan** paquetes concretos de las inferiores, solo sus interfaces. El único lugar donde se instancian implementaciones concretas es `cmd/server/main.go`.

---

## Auth — dos tokens independientes

| Aspecto        | Cliente (sesión QR)               | Admin                                                    |
| -------------- | --------------------------------- | -------------------------------------------------------- |
| Env var secret | `SESSION_SECRET`                  | `ADMIN_SECRET`                                           |
| Claims struct  | `auth.SessionClaims`              | `auth.AdminClaims`                                       |
| Expiración     | 30 minutos                        | 24 horas                                                 |
| Middleware     | `mw.AuthMiddleware(secret)`       | `mw.AdminAuthMiddleware(adminSecret)`                    |
| Contexto key   | `"session_claims"`                | `"admin_claims"`                                         |
| WS             | acepta ambos tokens vía `?token=` | `claims.RestaurantID` debe coincidir con `:restaurantId` |

`VerifyAdminSession` rechaza tokens de cliente aunque estén firmados con el mismo secret (valida `admin_id != ""`).

---

## RBAC — tres roles de admin

| Rol          | `RestaurantID` | Acceso                             |
| ------------ | -------------- | ---------------------------------- |
| `superadmin` | nil            | Todo el sistema                    |
| `owner`      | requerido      | Solo su restaurante                |
| `employee`   | requerido      | Rutas operativas de su restaurante |

`AdminClaims.Role` viaja en el JWT. Middlewares:

- `mw.RequireRole(roles...)` — 403 si el rol no está en la lista. Usar después de `AdminAuthMiddleware`.
- `mw.RequireRestaurantScope()` — superadmin pasa siempre; owner/employee deben coincidir con `:restaurantId`.

Lógica de permisos en el usecase (`AdminUseCase`):

- Owner solo puede crear/eliminar employees de su propio restaurante.
- `SeedAdminIfNeeded(repo)` crea un superadmin sin `restaurant_id` si no existe ningún admin.

`AdminUser.RestaurantID` es `*string` (puntero nullable) — nil para superadmin, obligatorio para owner/employee.

---

## Patrones obligatorios

### Mocks

`mocks/mocks.go` usa override functions opcionales:

```go
repo.CreateFn = func(r *entity.Restaurant) error { return someError }
```

Sin override → comportamiento in-memory por defecto. **Actualizar el mock en el mismo paso que la interfaz.**

### Errores de negocio

Definir errores centinela en el usecase y chequear con `errors.Is` en el handler:

```go
// usecase
var ErrTableNumberExists = errors.New("table number already exists")

// handler
if errors.Is(err, usecase.ErrTableNumberExists) {
    c.JSON(http.StatusConflict, ...)
}
```

### DTOs

Todos los request/response bodies viven en `interfaces/http/dto/dto.go`. No usar entidades de dominio directamente en handlers.

### SetupRouter

Firma actual (actualizar tests al cambiar):

```go
SetupRouter(requestH, feedbackH, restaurantH, sessionH, adminH, hub, secret, adminSecret)
```

---

## Comandos

```bash
# Build
go build ./...

# Tests (obligatorio antes de cerrar cualquier tarea)
go test ./...

# Servidor
go run cmd/server/main.go

# Swagger (tras modificar annotations en interfaces/http/)
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs
```

---

## Reglas de implementación obligatorias

### Build y tests tras cada cambio

Después de **cualquier** implementación ejecutar en orden hasta exit code 0:

```bash
go build ./...
go test ./...
```

### Contrato de no-regresión

- Cambio de firma → corregir **todos** los tests afectados en el mismo paso.
- Cambio de interfaz → actualizar `mocks/mocks.go` en el mismo paso.
- Cambio en `SetupRouter` o constructores de handlers → actualizar `router_test.go` en el mismo paso.

### Cobertura mínima por capa

| Capa                         | Qué testear                                                                     |
| ---------------------------- | ------------------------------------------------------------------------------- |
| `application/usecase`        | Caso feliz + errores de negocio (credenciales inválidas, duplicados, not found) |
| `interfaces/http`            | Código HTTP correcto para input válido e inválido                               |
| `infrastructure/auth`        | Sign + Verify: válido, expirado, firma incorrecta                               |
| `interfaces/http/middleware` | Token válido pasa · ausente/inválido → 401                                      |

### Cuándo agregar tests

- Handler nuevo → `*_handler_test.go` en el mismo paquete (crear si no existe).
- Usecase nuevo → `*_test.go` en el mismo paquete (agregar al existente si ya hay uno).
- Middleware nuevo → agregar casos a `middleware_test.go`.

---

## Swagger / swaggo

Cada handler HTTP en `interfaces/http/` **debe** tener annotations. Formato:

```go
// NombreFuncion godoc
// @Summary      Descripción corta
// @Tags         admin|requests|feedback|restaurants|tables|websocket
// @Accept       json
// @Produce      json
// @Param        nombre  path|body|query  tipo  requerido  "descripción"
// @Success      200  {object}  dto.TipoRespuesta
// @Failure      400  {object}  dto.ErrorResponse
// @Router       /api/v1/ruta [método]
```

- `@Tags` debe ser uno de: `admin`, `requests`, `feedback`, `restaurants`, `tables`, `websocket`
- `@Router` debe coincidir exactamente con la ruta en `router.go`
- Regenerar docs después de modificar annotations (ver comando arriba)

---

## Gotchas

- **GORM AutoMigrate + constraints existentes**: usar `unique` en lugar de `uniqueIndex` en GORM tags para columnas en tablas ya creadas en producción (evita DROP CONSTRAINT por nombre). Ver `entity.AdminUser`.
- **Admin seed**: `SeedAdminIfNeeded(repo)` crea un superadmin sin `restaurant_id` si no existe ningún admin. No requiere configuración extra.
- **`SignAdminSession` no sobreescribe `ExpiresAt`** si ya viene seteado en los claims — útil para tests con tokens expirados.
- **Rate limiter**: 20 req/min por IP, burst 20. Cleanup automático cada 10 min. IP extraída via `X-Forwarded-For`.
- **WS origin allowlist**: configurada vía `ALLOWED_ORIGINS` (comma-separated). En dev, si está vacío acepta todo.
- **`AdminUser.RestaurantID` es `*string`**: nil para superadmin, obligatorio para owner/employee. Acceder con `if admin.RestaurantID != nil { rid = *admin.RestaurantID }`.
- **Gin no permite rutas duplicadas**: si owner y superadmin comparten un endpoint, registrarlo una sola vez con `RequireRole(owner, superadmin)` y dejar que el handler/usecase diferencie la lógica.
