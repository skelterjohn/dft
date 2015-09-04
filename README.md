dft is a tool for (d)ata (f)iltering and (t)ransformation.

Data comes in on stdin, formatted as a json blob (for now), and comes outafter having had the filters and transformations applied.

`Usage: dft [FILTER|TRANSFORM]* [OUTPUT]`

Each filter and transform is applied to the entire object in the order they appear on the command line.

#examples#

The test files are meant to be read from top to bottom as tutorials. Start with `filter_test.go`, then `transform_test.go`, and finally `output_test.go`.

####filter Google Compute Engine instances by metadata key####

Real instance lists would have a lot more data, but I trimmed it down for readability (by humans...dft doesn't mind at all).
```
$ cat in.json 
[
  {
    "metadata": {
      "items": [
        {
          "key": "who",
          "value": "owned-by-jasmuth"
        },
        {
          "key": "startup-script",
          "value": "/root/start_worker.bash"
        }
      ]
    },
    "name": "process-1"
  },
  {
    "metadata": {
      "items": [
        {
          "key": "who",
          "value": "owned-by-someone-else"
        },
        {
          "key": "startup-script",
          "value": "/root/start_worker.bash"
        }
      ]
    },
    "name": "process-2"
  }
]
$ cat in.json | dft \
		'# args like this are comments' \
		'# first, filter out objects that do not have a who key' \
		'f:[].metadata.items[].key=who' \
		'# then copy that value to somewhere higher in the object' \
		't:[]{.who=.metadata.items[0].value}' \
		'# remove all fields but who and name' \
		'f:[]@who,name' \
		'# remove items that do not have my name in the foreman ID' \
		'f:[].who=/.*jasmuth/'
[
  {
    "foremanID": "foreman-not-on-borg-jasmuth",
    "name": "worker-ecba9d66-1c90-465c-8dd0-12e3ae867b66"
  }
]
```
