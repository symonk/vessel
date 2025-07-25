# A simple lightweight test server for development/debugging.
from aiohttp import web

async def handle(request):
    return web.json_response({"status": "ok"})

app = web.Application()
app.router.add_get("/", handle)

web.run_app(app, host="0.0.0.0", port=8000)