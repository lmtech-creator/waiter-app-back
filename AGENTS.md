# Waiter Backend — AGENTS.md

Go 1.25 · Gin · GORM/PostgreSQL · JWT HS256 · Zap · gorilla/websocket  
Módulo: `github.com/waiter/back`

## Arquitectura (Clean Architecture / DDD)

| Capa | Paquete | Rol |
|------|---------|-----|
| Entidades | `domain/entity/` | Structs puras, sin lógica |
| Repositorios | `domain/repository/` | Solo interfaces |
| Casos de uso | `application/usecase/` | Lógica de negocio; recibe interfaces |
| Handlers HTTP | `interfaces/http/` | Gin handlers + DTOs en `dto/` |
| Middleware | `interfaces/http/middleware/` | Auth, RBAC, rate limit |
| Persistencia | `infrastructure/persistence/` | GORM + PostgreSQL, AutoMigrate |
| Auth | `infrastructure/auth/` | JWT sign/verify |
| WebSocket | `infrastructure/websocket/` | Hub por `restaurantId` |
| Mocks | `mocks/mocks.go` | Un solo archivo; actualizar al cambiar interfaz |
| Wiring | `cmd/server/main.go` | DB → repos → usecases → handlers → router |

**Regla:** Capas superiores solo importan interfaces de inferiores. Solo `main.go` instancia implementaciones concretas.

## Auth — dos tokens independientes

| | Cliente (QR) | Admin |
|---|---|---|
| Secret env var | `SESSION_SECRET` | `ADMIN_SECRET` |
| Claims | `SessionClaims` | `AdminClaims` |
| Expiración | 30 min | 24 h |
| Middleware | `AuthMiddleware(secret)` | `AdminAuthMiddleware(adminSecret)` |
| Context key | `"session_claims"` | `"admin_claims"` |

`VerifyAdminSession` rechaza tokens de cliente aunque estén firmados con el mismo secret (valida `admin_id != ""`).

## RBAC

| Rol | `RestaurantID` | Acceso |
|-----|---------------|--------|
| `superadmin` | nil | Todo el sistema |
| `owner` | requerido | Solo su restaurante |
| `employee` | requerido | Rutas operativas |

Middlewares (usar en orden): `AdminAuthMiddleware → RequireRole(...) → RequireRestaurantScope()`

## Comandos

```bash
go build ./...            # Build
go test ./...             # Tests
go run cmd/server/main.go # Servidor (requiere .env)
# Swagger (tras modificar annotations en interfaces/http/)
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs
```

## Testing

- Mocks in-memory con `*Fn` override functions opcionales (`mocks/mocks.go`)
- Router tests usan `buildTestRouter()` en `router_test.go`
- Handler tests crean contexto con `gin.CreateTestContext(w)` e `injectClaims()`
- Secrets en tests: cadenas de ≥32 bytes

**Reglas de no-regresión:**
- Cambio de firma → corregir todos los tests en el mismo paso
- Cambio de interfaz → actualizar `mocks/mocks.go` en el mismo paso
- Cambio en `SetupRouter` o constructores → corregir `router_test.go` en el mismo paso

## Convenios de código

- DTOs: todos en `interfaces/http/dto/dto.go` (nunca usar entidades de dominio directamente en handlers)
- Errores de negocio: centinelas con `errors.Is` en handler, `errors.As` para `CooldownError`
- Swagger annotations obligatorias en cada handler (`@Tags` debe ser uno de: `admin`, `requests`, `feedback`, `restaurants`, `tables`, `websocket`)
- `SetupRouter` firma actual: `(requestH, feedbackH, restaurantH, sessionH, adminH, hub, secret, adminSecret) *gin.Engine`

## Datos

- `entity.JSONB` es un `[]string` con soporte `driver.Valuer`/`sql.Scanner` para columna JSONB en PostgreSQL. Usar `entity.JSONB(slice)` para convertir.
- Migraciones manuales en `migrations/`. AutoMigrate de GORM no altera columnas existentes; agregar `ALTER TABLE` manual en producción.

## Gotchas

- `AdminUser.RestaurantID` es `*string` (puntero nullable) — nil para superadmin, obligatorio owner/employee
- `SignAdminSession` no sobreescribe `ExpiresAt` si ya viene seteado — útil para tests con tokens expirados
- Rate limiter: 20 req/min por IP, burst 20, cleanup automático cada 5 min
- `ALLOWED_ORIGINS` vacío en dev → acepta todo (WS origin check)
- Dockerfile regenera Swagger en build (`swag init -g cmd/server/main.go -o docs`)
- Seed inicial: si no hay admins, crea `superadmin` con password generado y lo loguea
