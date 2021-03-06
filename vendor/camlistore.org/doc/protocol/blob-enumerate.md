# Blob Enumerate Protocol

The `/camli/enumerate-blobs` endpoint enumerates all blobs that the
server knows about.

They're returned in sorted order, sorted by (digest_type,
digest_value).  That is, md5-acbd18db4cc2f85cedef654fccc4a4d8 sorts
before sha1-0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33 because "m" sorts
before "s", even though "0" sorts before "a".

    GET /camli/enumerate-blobs?after=&limit= HTTP/1.1
    Host: example.com

URL GET parameters:

    after     optional    If provided, only blobs GREATER THAN this
                          value are returned.

                          Can't be used in combination with 'maxwaitsec'

    limit     optional    Limit the number of returned blobrefs.  The
                          server may have its own lower limit, however,
                          so be sure to pay attention to the presence
                          of a "continueAfter" key in the JSON response.

    maxwaitsec optional   The client may send this, an integer max
                          number of seconds the client is willing to
                          wait for the arrival of blobs.  If the
                          server supports long-polling (an optional
                          feature), then the server will return
                          immediately if any blobs or available, else
                          it will wait for this number of seconds.
                          It is an error to send this option with a non-
                          zero value along with the 'after' option.
                          The server's reply must include
                          "canLongPoll" set to true if the server
                          supports this feature.  Even if the server
                          supports long polling, the server may cap
                          'maxwaitsec' and wait for less time than
                          requested by the client.

                          Can't be used in combination with 'after'.


Response:

    HTTP/1.1 200 OK
    Content-Type: text/javascript

    {
      "blobs": [
        {"blobRef": "md5-acbd18db4cc2f85cedef654fccc4a4d8",
         "size": 3},
        {"blobRef": "sha1-0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33",
         "size": 3},
      ],
      "continueAfter": "sha1-0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33",
      "canLongPoll": true,
    }

Response keys:

    blobs          required   Array of {"blobRef": BLOBREF, "size": INT_bytes}
                              will be an empty list if no blobs are present.

    continueAfter  optional   If present, the result is truncated and there are
                              are (likely) more blobs after the provided
                              blobref, which should be passed to the next
                              request's "after" request parameter. It's possible
                              but rare that the final page of actual results has
                              continueAfter set, but the subsequent page is
                              empty. (if numBlobs % limit == 0)

    canLongPoll    optional   Set to true (type boolean) if the server supports
                              long polling.  If not true, the server ignores
                              the client's "maxwaitsec" parameter.
