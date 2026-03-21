defmodule RiceChatWeb.Endpoint do
  use Phoenix.Endpoint, otp_app: :rice_chat

  socket "/socket", RiceChatWeb.UserSocket,
    websocket: [check_origin: false],
    longpoll: false

  plug Plug.RequestId
  plug Plug.Telemetry, event_prefix: [:phoenix, :endpoint]
  plug Plug.Parsers,
    parsers: [:urlencoded, :multipart, :json],
    pass: ["*/*"],
    json_decoder: Jason

  plug CORSPlug
  plug RiceChatWeb.Router
end
