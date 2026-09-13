# Agent Log

This directory holds agent registration records and handoff records for the HomeService Meeting Setter.

## Structure

```
docs/agent-log/
├── README.md          (this file)
├── registrations/     (agent-registration.json records)
└── handoffs/          (handoff-record.json files)
```

## When to write

- **Registration:** At the start of any significant session. Use `templates/agent-registration.json`.
- **Handoff:** At the end of any significant session. Use `templates/handoff-record.json`.

## File naming

- Registrations: `YYYY-MM-DD-<agent-name>.json`
- Handoffs: `YYYY-MM-DD-<agent-name>-<topic>.json`

## Retention

Records are kept indefinitely for auditability. They are not gitignored — commit them so the team has full traceability.
