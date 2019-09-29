## Nic 0.2.1

+ add file handling method: nic.Response.SaveFile

+ change file uploading's API, now we can change file's MIME type and filename field

## Nic 0.2.0

+ refactor code
+ add the `Option` interface in order to be allowed pass both pointer and value parameters
+ fix upload files' bug
+ add the Chunked options, to contol whether `Transfer-Encoding: Chunked` is used

## Nic 0.1.2

+ Change nic.KV's type from `map[string]string` to `map[string]interface{}`, aim to support JSON serializer.
