<!-- SPDX-FileCopyrightText: 2026 Rillan AI LLC -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<!-- version: 3.3.0 -->
# Terraform Documentation

Use this mode to create or review module READMEs, root-module/operator guidance, or generated Terraform references. Document observed behavior from HCL, callers, workflows, backend configuration, tests, and existing repository conventions. Do not infer resources, defaults, permissions, or procedures from names alone.

## Discover before writing

Identify the Terraform/OpenTofu command and version, deployable root modules, child-module callers, provider and module constraints, lock files, backend/state boundaries, variable sources, outputs and consumers, test commands, CI entry points, and existing documentation generator. Treat secrets, local overrides, and uncommitted tfvars as evidence to handle carefully, never content to reproduce.

When documentation and code disagree, update the documentation to describe current behavior and call out any unsafe or likely unintended discrepancy. Do not quietly rewrite operational truth as an aspiration.

## Match the artifact to its reader

A reusable module README should explain:

- its purpose, ownership boundary, and resources or external data it manages;
- a minimal valid usage example using the repository's source/version convention;
- required Terraform/OpenTofu and provider versions;
- inputs, outputs, and consequential defaults;
- provider aliases or external prerequisites consumers must supply;
- identity, permissions, network exposure, destructive behavior, and sensitive outputs when relevant; and
- how to run the module's actual validation and tests.

A root-module or operator guide should add the working directory, state/workspace identity, variable sources, authentication method, dependencies on other state or systems, plan/apply approval flow, rollback or recovery boundaries, and emergency ownership. Describe credential variable names or identity flow without embedding secret values.

Keep examples short and executable. Demonstrate required wiring or a likely pitfall; omit exhaustive option catalogs. Link to provider documentation for volatile details. Put rationale for a significant, durable architecture choice in an ADR rather than expanding a README into decision history.

## Describe interfaces precisely

Every public variable and output needs a useful HCL `description`; generated tables can only be as accurate as those descriptions.

- For a variable, state what it controls, units or format, meaningful constraints, and the effect of its default. Avoid repeating the name.
- For a complex object, describe non-obvious attributes and optional behavior close to the type definition.
- For an output, state what it represents and any known cross-module or cross-state consumer. Mark secret-bearing values `sensitive`, while documenting that sensitive values remain in state.
- Document compatibility or breaking changes when variable requirements, types, defaults, output names, resource addresses, provider constraints, or state boundaries change.

## Generated sections

Discover and use the repository's pinned `terraform-docs` invocation: a task, Make target, pre-commit hook, CI command, container, or configuration-driven CLI command. Do not invent a command or replace hand-written prose.

When the repository uses injection mode, preserve prose outside its markers and regenerate only the marked content. `terraform-docs` injection replaces content between its configured comments; replace mode overwrites the target file. Review the resulting diff and require it to contain only changes explained by the HCL edit.

If no generator is configured, follow the repository's existing table style or propose a generator separately. Do not add a new tool dependency during an unrelated documentation change.

References: [`terraform-docs` usage](https://terraform-docs.io/user-guide/introduction/), [output and injection behavior](https://terraform-docs.io/user-guide/configuration/output/).

## Verification

Cross-check the finished document against the relevant `.tf` files and every affected caller. Run the discovered documentation command when available, then inspect the diff for stale inputs, outputs, defaults, requirements, examples, or generated sections. Validate example HCL when the repository provides a safe fixture or command.

If providers, modules, network, backend access, credentials, or tooling are unavailable, perform static cross-checks and state the exact verification gap. Documentation generation does not prove that a configuration validates, plans, or applies.
