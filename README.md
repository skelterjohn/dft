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
$ cat in.json | dft \
		'# args like this are comments' \
		'# first, filter out objects that do not have a foremanID key' \
		'f:[].metadata.items[].key=foremanID' \
		'# then copy that value to somewhere higher in the object' \
		't:[]{.foremanID=.metadata.items[0].value}' \
		'# remove all fields but foremanID and name' \
		'f:[]@foremanID,name' \
		'# remove items that do not have my name in the foreman ID' \
		'f:[].foremanID=/.*jasmuth/'
[
  {
    "foremanID": "foreman-not-on-borg-jasmuth",
    "name": "worker-ecba9d66-1c90-465c-8dd0-12e3ae867b66"
  }
]
```
