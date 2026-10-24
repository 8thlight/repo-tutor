# Architecture

The binary scans a repository, orders topics, and runs tutor turns through an agent driver. The core is pure functions over immutable values. Drivers and the entry point do input and output. The binary never calls a model API itself.

Components are not defined yet.
