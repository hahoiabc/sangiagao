import Config

config :rice_chat, RiceChatWeb.Endpoint,
  url: [host: "localhost"],
  render_errors: [formats: [json: RiceChatWeb.ErrorJSON]],
  pubsub_server: RiceChat.PubSub,
  live_view: [signing_salt: "rice_chat_salt"]

config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

config :phoenix, :json_library, Jason

import_config "#{config_env()}.exs"
