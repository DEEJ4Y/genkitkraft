import "@typespec/http";
import "./routes/health.tsp";
import "./routes/auth.tsp";
import "./routes/providers.tsp";
import "./routes/prompts.tsp";
import "./routes/agents.tsp";

using TypeSpec.Http;

@service(#{ title: "Hexagonal Arch Template API" })
@server("http://localhost:8080", "Local development")
namespace Api;
