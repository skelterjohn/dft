dft is a tool for (d)ata (f)iltering and (t)ransformation.

Data comes in on stdin, formatted as a json blob (for now), and comes out in the same format after having had the filters and transformations applied.

`Usage: dft [FILTER|TRANSFORM]*`

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
