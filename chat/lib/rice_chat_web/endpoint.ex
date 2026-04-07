defmodule RiceChatWeb.Endpoint do
  use Phoenix.Endpoint, otp_app: :rice_chat

  socket "/socket", RiceChatWeb.UserSocket,
    websocket: [
      check_origin: [
        "https://sangiagao.vn",
        "https://www.sangiagao.vn",
        "https://admin.sangiagao.vn"
      ],
      connect_info: [:x_headers]
    ],
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
