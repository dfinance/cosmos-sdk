# ADR 022: Custom baseapp panic handling

## Changelog

- 2020 Apr 23: Initial Draft

## Context

Currently [baseapp](https://github.com/cosmos/cosmos-sdk/blob/bad4ca75f58b182f600396ca350ad844c18fc80b/baseapp/baseapp.go#L55)
runTx method only [handles](https://github.com/cosmos/cosmos-sdk/blob/bad4ca75f58b182f600396ca350ad844c18fc80b/baseapp/baseapp.go#L538)
`OutOfGas` specific error (adding error context to it)
and provides a default handling mechanism (failing the Tx and logging an error).
That limits Cosmos SDK based project developers to add custom handling for their specific panic sources.

## Decision

We would like to design a mechanism to add custom handlers (middlewares) to `baseapp`'s `runTx()` panic processing.

## Status

Proposed

## Consequences

### Positive

- Developers of Cosmos SDK based projects can add custom panic handlers to:
    * add error context for custom panic sources (panic inside of custom keepers);
    * emit `panic()`: passthrough recovery object to the Tendermint core;
    * other necessary handling;
- Developers can use standard Cosmos SDK `baseapp` implementation, rather that injecting it to their project;

### Negative

- Introduces changes to the execution model design.

### Neutral



## References

- [PR-6053 with proposed solution](https://github.com/cosmos/cosmos-sdk/pull/6053)
