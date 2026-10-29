# 0470. License the project under GPL 3.0 or later

Status: accepted. Supersedes the all rights reserved wording in [0100](0100-about-page.md).

## Context

The repository carried an all rights reserved `LICENSE` until the company reviewed the project for release. The company has now cleared an open license, and the repository is moving under the company organization. The tool is a local binary that people install and run. Nobody uses it as a hosted service.

## Decision

License the project under the GNU General Public License, version 3 or later. The `LICENSE` file holds the unmodified license text. Eighth Light LLC stays the copyright holder, and the README says so.

Choose a strong copyleft license so that anyone who ships a modified binary must share their changes under the same terms. Choose version 3 over version 2 for its patent grant and because the Apache 2.0 code among the dependencies is compatible only with version 3. Choose it over the Affero license because a network clause protects nothing in a tool that runs on the learner's own machine. Every dependency uses a permissive license, so each one combines with GPL code, and `--version --notices` keeps shipping their notices.

## Consequences

The binary and its source can be copied, changed, and redistributed under the GPL. A distributed modified build must offer its source. A change of license later needs the agreement of every contributor, so the company should require a contributor agreement before it accepts outside changes.
