# 0550. Test ownership and the test layer

Status: accepted. The Java test directories also include `integration-test`, which a scan of a Spring MVC repository showed to be in use. A scan of Spring Petclinic showed its `@SpringBootTest` classes sit in the application's root package, so the Java rules now treat those as system tests even though a package name matches. This is the one case where "an owner wins over a system mark" does not hold.

## Context

A package is a directory, and a test file joins the package of its own directory. That works for Go, where `value_test.go` sits beside `value.go`, and for files that sit beside their code, such as JavaScript and TypeScript `.test.` and `.spec.` files and Python `test_*.py` files.

Other layouts keep tests in a mirror tree. Maven and Gradle put `src/test/java/com/acme/app/MainTest.java` in the package `java:src/test/java/com/acme/app`, apart from `java:src/main/java/com/acme/app`. A Python project often puts `tests/test_main.py` apart from `src/course/main.py`. The tutor builds a lesson from a package's files and its tests, but the production package has no tests, the test package has no files, and a package with no files is never a topic. The lesson for the code never sees its tests.

Tests that exercise the whole system have a second problem. A Spring Boot `@SpringBootTest` suite in a package such as `com.acme.integration` matches no production package and touches many at once. No single package owns it, and dropping it leaves a learner with no account of how the system is verified end to end. Other languages have the same shape, such as a Python `tests/integration/` directory.

The shared test rule also knows only the names `test_*`, `*_test.py`, `*.test.*`, and `*.spec.*`, and the directories `test`, `tests`, and `__tests__`. A test with another name or in another directory reads as production code. Its symbols appear in lessons and its imports add dependency edges that do not exist.

## Decision

A test is taught with the code it tests when one package owns it, and as part of a test layer when it exercises the system as a whole. A test that is neither stays where it is.

**Owner.** An analyzer may supply a rule that names the production package a test exercises. After the analyzer builds its index, the shared driver moves each owned test into that package's tests. A package left with no files and no tests is dropped.

**Test layer.** An analyzer may also mark a test as a system test. A system test with no owner makes its own directory's package a test layer. A test layer is a topic although it has no files. Its dependencies are the production packages its tests import, so the course orders it after the code it exercises. No production package depends on a test layer. Its lesson teaches how the system is tested end to end, from the tests and the code they reach.

**Neither.** A test with no owner that is not a system test stays in its own directory's package, which is not a topic.

**Order of rules.** An owner wins over a system mark. A test in the mirror of a package is a test of that package, even when it starts the whole application. Only a system test that no package claims forms a layer.

**Recognition.** An analyzer may supply a rule that recognizes a test file, in place of the shared rule. A language that supplies none keeps the shared rule.

A test adds no dependency edge to a production package, as today. Moving a test changes which lesson shows it and nothing in the order of production topics. An analyzer's rules claim only what a pattern can support, and a test they cannot place stays unowned, as [0530](0530-language-analyzers.md) drops a link it would have to guess.

## Java rules

- A file is a test when its name ends in `Test` or `Tests`, or in `IT` after a lowercase letter or a digit, or when its path has a directory named `test`, `integrationTest`, or `androidTest`.
- A test that names one class with `@WebMvcTest(X.class)` or `@WebFluxTest(X.class)` is owned by the package that declares `X`.
- Otherwise a test is owned by the package whose declared name matches the package the test declares. The directory does not matter. `package com.acme.app;` in a test pairs with the production package that declares `com.acme.app`. The rule names an owner only when exactly one production directory declares that package.
- A test annotated `@SpringBootTest` that has no owner is a system test.
- A `@SpringBootTest` in the package that declares a class annotated `@SpringBootApplication` has no owner by package name. Spring starts the whole application from that package, so the test exercises the system and not that package alone. It is a system test, and its own directory becomes a test layer. A `@SpringBootTest` in any other package keeps the owner its package name gives.

Frameworks other than Spring Boot are not recognized until a repository needs them.

## Other languages

Python and JavaScript and TypeScript keep the shared rules until a later decision gives each its own. Python is the next analyzer to get them. The direction, not yet decided:

- An import names the owner better than a name does, because a Python test must import what it tests. A test that imports exactly one production package would be owned by it.
- A `tests` directory inside a package, as Django apps have, would belong to the parent package.
- Django's `tests.py` would be recognized as a test file.
- System-test signals include `TestClient(` from FastAPI, `test_client()` from Flask, Django's `Client` and `LiveServerTestCase`, and the pytest marks `integration` and `e2e`. A test that imports several production packages and has no owner is a weaker signal.

JavaScript and TypeScript would need rules of their own, since a `.test.` file that sits beside its code is already owned.

## Consequences

The lesson for a Java package includes its tests, so a learner is taught how the code is checked as well as what it does. A Spring repository gains one topic per test layer, taught after the packages it exercises. Python and JavaScript and TypeScript lessons do not change until their rules land, and a test in a mirror directory there still goes untaught.

Topic identifiers for production packages do not change. Existing sessions gain topics and lose none, and the session file shape does not change, so the format version stays as it is under [0460](0460-session-format-policy.md).

The Java scan warning stops saying that tests outside `src/test` read as source, and names what it still misses: a `@SpringBootTest` split across lines or built from a constant, and system tests that use another framework. The tutor's lesson instructions must say what a test layer is. The architecture document describes the ownership rule and the layer once the code follows them.

The work lands as separate commits: the driver hooks for recognition and ownership, the Java recognition rule, Java ownership by package name, Java ownership by a named class, and the test layer. The layer needs the scan to read imports in test files, which it skips today.
