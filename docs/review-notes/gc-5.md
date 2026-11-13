# gc-5: Cache-Control on GET /api/version

GET /api/version now sends `Cache-Control: public, max-age=60`, because its answer only changes with a deploy and clients may cache it briefly. Other API routes are unchanged. The handler test now checks the header.
