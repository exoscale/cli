# go-jmespath-plus - A JMESPath implementation in Go

![CI](https://github.com/danielgtaylor/go-jmespath-plus/workflows/CI/badge.svg?branch=master)

See http://jmespath.org for more info.

## Enhancements

### Recursive Descent Operator

A new `..` recursive descent operator is supported. It will recursively traverse all objects and arrays to create a new projection. For example, given:

```json
{
  "foo": {
    "bar": 1,
    "baz": [
      {
        "bar": 2
      },
      {
        "bar": 3
      }
    ]
  }
}
```

Then you can see the difference between the a wildcard and recursive descent:

```sh
# Wildcard key, selects only the first `bar`.
$ jpgo '*.bar' <input
[1]

# Recursive descent, selects *all* the `bar` entries.
$ jpgo '..bar' <input
[1, 2, 3]
```

### Multi-select Hash Shorthand

Selecting a single key from an object is now possible with a property shorthand syntax borrowed from [modern Javascript](http://es6-features.org/#PropertyShorthand). Given:

```json
{
  "foo": {
    "bar": 1,
    "baz": [true, false]
  }
}
```

Then you can select via the shorthand:

```sh
# Shorthand example.
$ jpgo 'foo.{bar}' <input
{"bar": 1}

# Mixing shorthand and long-form.
$ jpgo 'foo.{bar, first_baz: baz[0]}' <input
{"bar": 1, "first_baz": true}
```

### Group By Function

The [`group_by`](https://stedolan.github.io/jq/manual/) function from `jq` is borrowed to generate a list of grouped objects based on the result of an expression executed on each item in the incoming array. The output is sorted in ascending order.

`group_by(array $elements, expression->number|expression->string field)`

Given:

```json
{
  "foo": [
    {
      "id": 1,
      "type": "red"
    },
    {
      "id": 2,
      "type": "blue"
    },
    {
      "id": 3,
      "type": "red"
    }
  ]
}
```

Then you can group the items by their `type`:

```sh
# Group the inputs by the type of item.
$ jpgo 'group_by(foo, &type)' <input
```

The result:

```json
[
  [
    {
      "id": 2,
      "type": "blue"
    }
  ],
  [
    {
      "id": 1,
      "type": "red"
    },
    {
      "id": 3,
      "type": "red"
    }
  ]
]
```

### Pivot Function

The `pivot` function is a convenience wrapper around `group_by(...)` that also pivots the data to return an object where the keys are the grouping value and the values are the groups of objects with the given projection expression applied to each object.

`pivot(array $elements, expression->number|expression->string field, expression projection)`

Given:

```json
{
  "foo": [
    {
      "id": 1,
      "type": "red"
    },
    {
      "id": 2,
      "type": "blue"
    },
    {
      "id": 3,
      "type": "red"
    }
  ]
}
```

Then you can pivot the items by their `type`:

```sh
# Pivot items by their type
$ jpgo 'pivot(foo, &type, &id)' <input
```

The result:

```json
{
  "blue": [2],
  "red": [1, 3]
}
```

You can also use the identity to keep the full objects:

```sh
# Pivot items by their type and keep each original item.
$ jpgo 'pivot(foo, &type, &@)' <input
```

The result:

```json
{
  "blue": [
    {
      "id": 2,
      "type": "blue"
    }
  ],
  "red": [
    {
      "id": 1,
      "type": "red"
    },
    {
      "id": 3,
      "type": "red"
    }
  ]
}
```
