# Changelog #

## master ##

## v3.3 ##
  * Introduce new `Num` and `OraNum` data types and `N` Go Column Type to represent Oracle numbers fully, by exchanging them with Go as strings.
    Make `N` the default column type for numbers unrepresentable by int64 or float64 (more digits than 19 or 15).

## v3.2 ##
  * Rewrite for Go 1.6 cgo restrictions.

## v3.1 ##

  * Use the returned length info for several places when reading attributes - such as column names.
  * Use float64 for non-integer numeric column types.

## v3.0 ##

  * Big rewrite of handles to pool them properly.
