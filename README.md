# trisse-home

A self-hosted personal start page in the spirit of iGoogle: freely arranged
widgets plus small modules for everyday life and the home server — as lean
as possible, as much as necessary, built on a strict Clean Architecture.

> Status: under active development. See the full documentation below once
> all modules are complete.

## Quick start

```sh
cp .env.example .env
sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
sudo docker compose up -d --build
sudo docker compose logs trisse-home | grep setup_token
```

Open `http://<server-ip>:8090`, enter the setup token from the log and choose
a password.

## License

[MIT](LICENSE)
