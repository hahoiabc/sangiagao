import Config

config :rice_chat, RiceChatWeb.Endpoint,
  http: [ip: {0, 0, 0, 0}, port: 4000],
  check_origin: true,
  server: true,
  secret_key_base: System.get_env("SECRET_KEY_BASE")
