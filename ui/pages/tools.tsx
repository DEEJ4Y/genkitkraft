import { useState } from "react";
import {
  Title,
  Text,
  Stack,
  Group,
  Button,
  Loader,
  Alert,
  Center,
  Pagination,
  Tabs,
} from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchClient } from "../lib/api/client";
import { HttpToolCard } from "../components/HttpToolCard";
import { HttpToolForm } from "../components/HttpToolForm";
import { McpServerCard } from "../components/McpServerCard";
import { McpServerForm } from "../components/McpServerForm";
import type { components } from "../lib/api/schema";

type HttpToolResponse = components["schemas"]["Models.HttpToolResponse"];
type McpServerResponse = components["schemas"]["Models.McpServerResponse"];

type View =
  | { mode: "list" }
  | { mode: "create" }
  | { mode: "edit"; toolId: string };

const PAGE_SIZE = 20;

function HttpToolsTab() {
  const queryClient = useQueryClient();
  const [view, setView] = useState<View>({ mode: "list" });
  const [page, setPage] = useState(1);

  const offset = (page - 1) * PAGE_SIZE;

  const toolsQuery = useQuery({
    queryKey: ["get", "/api/v1/http-tools", { limit: PAGE_SIZE, offset }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET("/api/v1/http-tools", {
        params: { query: { limit: PAGE_SIZE, offset } },
      });
      if (error) throw new Error("Failed to fetch HTTP tools");
      return data;
    },
  });

  const editingToolId = view.mode === "edit" ? view.toolId : null;

  const editingToolQuery = useQuery({
    queryKey: ["get", "/api/v1/http-tools", editingToolId],
    queryFn: async () => {
      if (!editingToolId) return null;
      const { data, error } = await fetchClient.GET("/api/v1/http-tools/{id}", {
        params: { path: { id: editingToolId } },
      });
      if (error) throw new Error("Failed to fetch HTTP tool");
      return data;
    },
    enabled: !!editingToolId,
  });

  function handleSaved() {
    setView({ mode: "list" });
    queryClient.invalidateQueries({ queryKey: ["get", "/api/v1/http-tools"] });
  }

  async function handleDelete(tool: HttpToolResponse) {
    if (!confirm(`Delete "${tool.name}"? This cannot be undone.`)) return;
    await fetchClient.DELETE("/api/v1/http-tools/{id}", {
      params: { path: { id: tool.id } },
    });
    queryClient.invalidateQueries({ queryKey: ["get", "/api/v1/http-tools"] });
  }

  if (view.mode === "create") {
    return (
      <HttpToolForm
        onSaved={handleSaved}
        onCancel={() => setView({ mode: "list" })}
      />
    );
  }

  if (view.mode === "edit") {
    if (editingToolQuery.isPending) {
      return (
        <Center py="xl">
          <Loader />
        </Center>
      );
    }

    if (editingToolQuery.error || !editingToolQuery.data) {
      return (
        <Alert color="red" variant="light">
          Failed to load HTTP tool.
        </Alert>
      );
    }

    return (
      <HttpToolForm
        tool={editingToolQuery.data}
        onSaved={handleSaved}
        onCancel={() => setView({ mode: "list" })}
      />
    );
  }

  const tools = toolsQuery.data?.httpTools ?? [];
  const total = toolsQuery.data?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <>
      <Group justify="space-between" align="center" mb="md">
        <div />
        <Button
          leftSection={<IconPlus size={16} />}
          onClick={() => setView({ mode: "create" })}
        >
          New HTTP Tool
        </Button>
      </Group>

      {toolsQuery.isPending && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {toolsQuery.error && (
        <Alert color="red" variant="light" mb="md">
          Failed to load HTTP tools.
        </Alert>
      )}

      {!toolsQuery.isPending && tools.length === 0 && (
        <Text c="dimmed" ta="center" py="xl">
          No HTTP tools yet. Create your first HTTP tool to get started.
        </Text>
      )}

      <Stack gap="sm">
        {tools.map((tool) => (
          <HttpToolCard
            key={tool.id}
            tool={tool}
            onEdit={() => setView({ mode: "edit", toolId: tool.id })}
            onDelete={() => handleDelete(tool)}
          />
        ))}
      </Stack>

      {totalPages > 1 && (
        <Center mt="lg">
          <Pagination total={totalPages} value={page} onChange={setPage} />
        </Center>
      )}
    </>
  );
}

type McpView =
  | { mode: "list" }
  | { mode: "create" }
  | { mode: "edit"; serverId: string };

function McpServersTab() {
  const queryClient = useQueryClient();
  const [view, setView] = useState<McpView>({ mode: "list" });
  const [page, setPage] = useState(1);

  const offset = (page - 1) * PAGE_SIZE;

  const serversQuery = useQuery({
    queryKey: ["get", "/api/v1/mcp-servers", { limit: PAGE_SIZE, offset }],
    queryFn: async () => {
      const { data, error } = await fetchClient.GET("/api/v1/mcp-servers", {
        params: { query: { limit: PAGE_SIZE, offset } },
      });
      if (error) throw new Error("Failed to fetch MCP servers");
      return data;
    },
  });

  const editingServerId = view.mode === "edit" ? view.serverId : null;

  const editingServerQuery = useQuery({
    queryKey: ["get", "/api/v1/mcp-servers", editingServerId],
    queryFn: async () => {
      if (!editingServerId) return null;
      const { data, error } = await fetchClient.GET(
        "/api/v1/mcp-servers/{id}",
        {
          params: { path: { id: editingServerId } },
        },
      );
      if (error) throw new Error("Failed to fetch MCP server");
      return data;
    },
    enabled: !!editingServerId,
  });

  function handleSaved() {
    setView({ mode: "list" });
    queryClient.invalidateQueries({ queryKey: ["get", "/api/v1/mcp-servers"] });
  }

  async function handleDelete(server: McpServerResponse) {
    if (!confirm(`Delete "${server.name}"? This cannot be undone.`)) return;
    await fetchClient.DELETE("/api/v1/mcp-servers/{id}", {
      params: { path: { id: server.id } },
    });
    queryClient.invalidateQueries({ queryKey: ["get", "/api/v1/mcp-servers"] });
  }

  if (view.mode === "create") {
    return (
      <McpServerForm
        onSaved={handleSaved}
        onCancel={() => setView({ mode: "list" })}
      />
    );
  }

  if (view.mode === "edit") {
    if (editingServerQuery.isPending) {
      return (
        <Center py="xl">
          <Loader />
        </Center>
      );
    }

    if (editingServerQuery.error || !editingServerQuery.data) {
      return (
        <Alert color="red" variant="light">
          Failed to load MCP server.
        </Alert>
      );
    }

    return (
      <McpServerForm
        server={editingServerQuery.data}
        onSaved={handleSaved}
        onCancel={() => setView({ mode: "list" })}
      />
    );
  }

  const servers = serversQuery.data?.mcpServers ?? [];
  const total = serversQuery.data?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <>
      <Group justify="space-between" align="center" mb="md">
        <div />
        <Button
          leftSection={<IconPlus size={16} />}
          onClick={() => setView({ mode: "create" })}
        >
          New MCP Server
        </Button>
      </Group>

      {serversQuery.isPending && (
        <Center py="xl">
          <Loader />
        </Center>
      )}

      {serversQuery.error && (
        <Alert color="red" variant="light" mb="md">
          Failed to load MCP servers.
        </Alert>
      )}

      {!serversQuery.isPending && servers.length === 0 && (
        <Text c="dimmed" ta="center" py="xl">
          No MCP servers configured yet. Add your first MCP server to get
          started.
        </Text>
      )}

      <Stack gap="sm">
        {servers.map((server) => (
          <McpServerCard
            key={server.id}
            server={server}
            onEdit={() => setView({ mode: "edit", serverId: server.id })}
            onDelete={() => handleDelete(server)}
          />
        ))}
      </Stack>

      {totalPages > 1 && (
        <Center mt="lg">
          <Pagination total={totalPages} value={page} onChange={setPage} />
        </Center>
      )}
    </>
  );
}

export default function ToolsPage() {
  return (
    <>
      <Title order={2} mb="lg">
        Tools
      </Title>
      <Text size="sm" c="dimmed" mb="lg">
        Define tools that agents can use during conversations via LLM function
        calling.
      </Text>

      <Tabs defaultValue="http">
        <Tabs.List mb="md">
          <Tabs.Tab value="http">HTTP Tools</Tabs.Tab>
          <Tabs.Tab value="mcp">MCP Servers</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="http">
          <HttpToolsTab />
        </Tabs.Panel>

        <Tabs.Panel value="mcp">
          <McpServersTab />
        </Tabs.Panel>
      </Tabs>
    </>
  );
}
