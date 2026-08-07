# note vs notes — Hibernate column mismatch

Status: resolved

The entity field `notes` mapped to a column Hibernate expected as `notes`, while the
Liquibase-created column was `note`. `ddl-auto: validate` failed at startup with
`missing column [notes]`.

Fixed by [[decisions/liquibase-over-column-alias]].

Related: [[concepts/hibernate]]
