# 0560. Python and JavaScript test rules

Status: accepted. It supersedes the "Other languages" section of [0550](0550-test-ownership-and-layer.md).

## Context

[0550](0550-test-ownership-and-layer.md) decides that a test joins the package it tests, or a test layer when it exercises the system. Java supplies the rules that name an owner and a system test. Python and JavaScript and TypeScript keep the shared rules, so a test in a mirror tree such as `tests/` goes untaught, and a test with a name the shared rule does not know reads as production code.

The principle is settled. What each language needs is the evidence a text pattern can support, because [0530](0530-language-analyzers.md) drops a link it would have to guess. A Java test declares its package, and that names its owner. A Python or JavaScript test declares nothing, so the evidence is where the file sits and what it imports.

## Decision

Python and JavaScript and TypeScript supply the same two ownership rules, in this order, and each supplies its own recognition and system-test signals.

**Directory rule.** A test directory nested inside a production package belongs to that package. A directory named `test`, `tests`, or `__tests__` whose parent directory is a production package gives its tests to the parent. Django's `app/tests/` and a `src/__tests__/` follow it. A directory at the repository root is a mirror of the whole tree and not nested in a package, so its tests go to the import rule. The parent must hold production files of the same language.

**Import rule.** Otherwise a test is owned by the production package that all its resolved local imports reach, when they reach exactly one. Imports of a test-only package, such as a fixtures directory, do not count. A test that imports two production packages, or none, has no owner. A test that imports many packages is as likely a unit test of a helper as an end-to-end test, so the number of imports is not a signal.

**Python recognition.** A file is a test when it matches the shared names or is named `tests.py`.

**Python system signals.** A test with no owner is a system test when it uses `TestClient(`, `test_client()`, `LiveServerTestCase`, or Django's `Client` imported from `django.test`, or carries the `integration` or `e2e` pytest mark, or sits in a directory named `integration` or `e2e`.

**JavaScript recognition.** A file is a test when it matches the shared names or ends in `.e2e-spec.` or `.e2e.` followed by an extension.

**JavaScript system signals.** A test with no owner is a system test when it imports `supertest`, `@playwright/test`, or `cypress`, or sits in a directory named `e2e` or `integration`.

**Order.** The order of 0550 holds. An owner wins over a system mark, and Java's `@SpringBootApplication` case stays Java's. The cost is that a `TestClient` test that imports only the package holding the app is taught with that package and not as a test layer. A later decision can add an application-package rule, as Java has one, if a scan of a real repository shows it.

Other frameworks are not recognized until a repository needs them. The scan warning for each language names what it still misses: dynamic imports, aliases, re-exports, and test helpers that hide the signal.

## Consequences

A Python or JavaScript lesson for a package includes the tests that sit in a mirror tree, and a repository with an `integration` or `e2e` suite gains a test layer. A test the rules cannot place stays where it is, as today. Production topic identifiers and their order do not change. Existing sessions gain topics and lose none, and the session file shape does not change, so the format version stays as it is under [0460](0460-session-format-policy.md).

The work lands as separate commits: the shared directory rule, the import rule, Python recognition and system signals, then JavaScript recognition and system signals. The architecture document and the scan warnings follow each commit. Test files already sit in the scan, and the import rule reuses the resolver that production imports use.
