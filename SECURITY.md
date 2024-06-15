# Security Policy

## Supported Versions

We support the latest release of SAGE and the current master branch.

If you find a security issue in an older release, please still report it to us,
to force a software upgrade to all affected users.

## Reporting a Vulnerability

Please report security vulnerabilities to us by using our
[contact methods](https://sage.party/contact).

We will respond to security reports within 24 hours, and will keep you informed
of our progress in fixing the vulnerability.

## Scope

This security policy applies to the SAGE software and all of its components, for
example:

- bypassing file encryption
- any XSS inside the UI
- any potential abuse of the RPC interface, like:
  - stealing the user's private key
  - accessing the user's files (except in the SAGE directory)
  - native code execution
