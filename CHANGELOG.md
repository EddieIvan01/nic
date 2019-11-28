## Nic 0.3.1

+ Add request/response hook function
+ Change testing cases

## Nic 0.3

+ Fix issues #5 #6
+ Add enhancement issue #2: add option params 
+ Change:  `nic.Session` is go-routine safe now

## Nic 0.2.1

+ Add file handling method: nic.Response.SaveFile
+ Change file uploading's API, now we can change file's MIME type and filename field

## Nic 0.2.0

+ Refactor code
+ Add the `Option` interface in order to be allowed pass both pointer and value parameters
+ Fix upload files' bug
+ Add the Chunked options, to contol whether `Transfer-Encoding: Chunked` is used

## Nic 0.1.2

+ Change nic.KV's type from `map[string]string` to `map[string]interface{}`, aim to support JSON serializer.
