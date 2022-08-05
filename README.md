# brock

brock is a go framework used internally within onebrick.io, this framework is
totally opinionated, and following only the convention within onebrick.io

## ErrXXX
the `localError` is a constant string and used as a reference to check the error
from the caller side

## HTTP
the `HTTP` variable is a namespace in which brock maintained

## Parser
brock unify the parser into one package, covers XML, JSON, YAML & TOML

## SQL
