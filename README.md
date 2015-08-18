dft is a tool for (d)ata (f)iltering and (t)ransformation.

Data comes in on stdin, formatted as a json blob (for now), and comes out in the same format after having had the filters and transformations applied.

`Usage: dft [FILTER|TRANSFORM]*`

Each filter and transform is applied to the entire object in the order they appear on the command line.

#examples#

####filter Google Compute Engine instances by metadata key####

Real instance lists would have a lot more data, but I trimmed it down for readability (by humans...dft doesn't mind at all).
```
$ cat in.json 
[
  {
    "metadata": {
      "items": [
        {
          "key": "foremanID",
          "value": "foreman-not-on-borg-jasmuth"
        },
        {
          "key": "startup-script",
          "value": "/root/start_worker.bash"
        }
      ]
    },
    "name": "worker-ecba9d66-1c90-465c-8dd0-12e3ae867b66"
  },
  {
    "metadata": {
      "items": [
        {
          "key": "foremanID",
          "value": "cloud-build-dev/devel.foreman.server/vn/0"
        },
        {
          "key": "startup-script",
          "value": "/root/start_worker.bash"
        }
      ]
    },
    "name": "worker-f4b6f6b1-088b-4c14-ae5e-2b31c0cbe305"
  }
]
$ cat in.json | dft 'f:[].metadata.items[].key=foremanID' 't:[]{.metadata.items[0].value=.foremanID}' 'f:[]@foremanID,name' 'f:[].foremanID=.*jasmuth'
[
  {
    "foremanID": "foreman-not-on-borg-jasmuth",
    "name": "worker-ecba9d66-1c90-465c-8dd0-12e3ae867b66"
  }
]
```

#specification#

##filters##

A filter begins with the string `f:`, and then has operators that dig into the object and compare values against provided input.

###filter operators###

```[<index>]``` - list index. Match if the rest of the filter matches this item.

```[]``` - list "only". All items in the list will be filtered out if they do not match the rest of the filter.

```[E]``` - list "any". If at least one of the indices matches the rest of the filter, then no indices are filtered out.

```.<field>``` - object field. Match if the rest of the filter matches this item. If no operations follow, this will be a filter on the existence of the field.

```.()``` - field "only". All fields in the object will be filtered out if they do not match the rest of the filter.

```.(E)``` - field "any". If at least one of the fields matches the rest of the filter, then no fields are filtered out. 

```=<value>``` - explicit match, if the item matches the value. May be a regular expression for strings. No operations may follow.

```@<item1>,<item2>,...,<item3>``` - item cut. Filter all items whose names are not listed. May be indices or field names. No operations may follow.

```{<op1>,<op2>,...,<op3>}``` - filter intersection. Match if the item matches all the subfilters. No operations may follow.

##transforms##

A transform begins with the string `t:`, has some operators for digging down into the object, and then a "get" and "set" expression surrounded by braces.

###transform operators###

```[<index>]``` - list index. Apply the remainder of the transform to the indexed item.

```[]``` - list "all". Apply the remainder of the transform to all items in the list.

```.<field>``` - struct field. Apply the remainder of the transform to the field.

```.()``` - struct "all". Apply the remainder of the transform to all fields in the structure.

```{<get>,<set>}``` - value replace. The "get" expression is like a general transform operator, except `[]` and `.()` are disallowed. The "set" expression is either `.<field>` or `[<index>]`, with nothing else.
