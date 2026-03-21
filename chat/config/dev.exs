import Config

config :rice_chat, RiceChatWeb.Endpoint,
  http: [ip: {0, 0, 0, 0}, port: 4000],
  check_origin: false,
  debug_errors: true,
  secret_key_base: "dev-secret-key-base-at-least-64-bytes-long-for-development-only-change-in-prod"
