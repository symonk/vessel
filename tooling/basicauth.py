from aiohttp import web
import base64

USERNAME = "foo"
PASSWORD = "foo"

def check_basic_auth(request):
    auth_header = request.headers.get('Authorization')
    if not auth_header or not auth_header.startswith('Basic '):
        return False

    try:
        encoded = auth_header.split(' ')[1]
        decoded = base64.b64decode(encoded).decode('utf-8')
        user, pwd = decoded.split(':', 1)
        return user == USERNAME and pwd == PASSWORD
    except Exception:
        return False

async def protected_handler(request):
    if not check_basic_auth(request):
        return web.Response(status=401, headers={
            'WWW-Authenticate': 'Basic realm="Access to the site"',
        }, text='Unauthorized')

    return web.Response(text="Welcome, authorized user!")

app = web.Application()
app.add_routes([web.get('/secure', protected_handler)])

web.run_app(app)
