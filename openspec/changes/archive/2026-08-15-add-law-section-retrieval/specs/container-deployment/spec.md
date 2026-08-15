## ADDED Requirements

### Requirement: Contenedor endurecido

Las configuraciones de despliegue provistas (`docker-compose.yml` y los targets podman del Makefile) DEBEN ejecutar el contenedor endurecido: sistema de archivos raíz de solo lectura (con un tmpfs para `/tmp`), capacidades Linux mínimas (`cap_drop: ALL`) y `no-new-privileges`. El servidor NO DEBE necesitar escritura en disco para operar.

#### Scenario: Rootfs de solo lectura
- **WHEN** el contenedor se ejecuta con la configuración provista
- **THEN** el sistema de archivos raíz es read-only y el servidor arranca y responde sin escrituras a disco

#### Scenario: Capacidades mínimas
- **WHEN** el contenedor se ejecuta con la configuración provista
- **THEN** el contenedor corre con `cap_drop: ALL` y `no-new-privileges: true`
