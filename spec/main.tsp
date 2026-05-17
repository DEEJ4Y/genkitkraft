import "@typespec/http";
import "./routes/health.tsp";
import "./routes/auth.tsp";
import "./routes/providers.tsp";
import "./routes/prompts.tsp";
import "./routes/agents.tsp";
import "./routes/playground.tsp";
import "./routes/deploy.tsp";
import "./routes/http_tools.tsp";
import "./routes/mcp_servers.tsp";
import "./routes/agent_tools.tsp";
import "./routes/built_in_tools.tsp";

using TypeSpec.Http;

@service(#{ title: "Hexagonal Arch Template API" })
@server("http://localhost:8080", "Local development")
namespace Api;
