import { useState } from "react";
import {
  TextInput,
  Textarea,
  Select,
  Button,
  Group,
  Stack,
  Alert,
  Text,
  ActionIcon,
  Checkbox,
  Card,
} from "@mantine/core";
import {
  IconArrowLeft,
  IconBraces,
  IconPlus,
  IconTrash,
} from "@tabler/icons-react";
import { fetchClient } from "../lib/api/client";
import type { components } from "../lib/api/schema";

type HttpToolResponse = components["schemas"]["Models.HttpToolResponse"];

interface HttpToolFormProps {
  tool?: HttpToolResponse;
  onSaved: () => void;
  onCancel: () => void;
}

interface HeaderRow {
  name: string;
  value: string;
}

interface ParamRow {
  name: string;
  type: string;
  description: string;
  required: boolean;
}

const HTTP_METHODS = [
  { value: "GET", label: "GET" },
  { value: "POST", label: "POST" },
  { value: "PUT", label: "PUT" },
  { value: "DELETE", label: "DELETE" },
  { value: "PATCH", label: "PATCH" },
  { value: "HEAD", label: "HEAD" },
  { value: "OPTIONS", label: "OPTIONS" },
];

const PARAM_TYPES = [
  { value: "string", label: "string" },
  { value: "number", label: "number" },
  { value: "integer", label: "integer" },
  { value: "boolean", label: "boolean" },
];

function parseInputSchema(schema: string): ParamRow[] {
  try {
    const parsed = JSON.parse(schema);
    if (!parsed.properties) return [];
    const required: string[] = parsed.required ?? [];
    return Object.entries(parsed.properties).map(
      ([name, prop]: [string, any]) => ({
        name,
        type: prop.type ?? "string",
        description: prop.description ?? "",
        required: required.includes(name),
      }),
    );
  } catch {
    return [];
  }
}

function buildInputSchema(params: ParamRow[]): string {
  if (params.length === 0) return "{}";
  const properties: Record<string, any> = {};
  const required: string[] = [];
  for (const p of params) {
    if (!p.name.trim()) continue;
    properties[p.name.trim()] = {
      type: p.type,
      ...(p.description ? { description: p.description } : {}),
    };
    if (p.required) required.push(p.name.trim());
  }
  if (Object.keys(properties).length === 0) return "{}";
  return JSON.stringify({
    type: "object",
    properties,
    ...(required.length > 0 ? { required } : {}),
  });
}

type HttpMethod =
  | "GET"
  | "POST"
  | "PUT"
  | "DELETE"
  | "PATCH"
  | "HEAD"
  | "OPTIONS";

export function HttpToolForm({ tool, onSaved, onCancel }: HttpToolFormProps) {
  const isEdit = !!tool;
  const [name, setName] = useState(tool?.name ?? "");
  const [description, setDescription] = useState(tool?.description ?? "");
  const [method, setMethod] = useState<HttpMethod>(
    (tool?.method as HttpMethod) ?? "GET",
  );
  const [url, setUrl] = useState(tool?.url ?? "");
  const [headers, setHeaders] = useState<HeaderRow[]>(
    tool?.headers?.map((h) => ({ name: h.name, value: h.value })) ?? [],
  );
  const [bodyTemplate, setBodyTemplate] = useState(tool?.bodyTemplate ?? "");
  const [params, setParams] = useState<ParamRow[]>(
    tool ? parseInputSchema(tool.inputSchema) : [],
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [urlError, setUrlError] = useState("");

  function validateUrl(value: string): string {
    if (!value.trim()) return ""
    // Replace Go template expressions with a placeholder so URL constructor can parse
    const sanitized = value.replace(/\{\{.*?\}\}/g, "PLACEHOLDER")
    try {
      new URL(sanitized)
      return ""
    } catch {
      return "Invalid URL. Must include a scheme (e.g. https://)"
    }
  }

  function addHeader() {
    setHeaders([...headers, { name: "", value: "" }]);
  }

  function removeHeader(index: number) {
    setHeaders(headers.filter((_, i) => i !== index));
  }

  function updateHeader(index: number, field: "name" | "value", value: string) {
    const updated = [...headers];
    updated[index] = { ...updated[index], [field]: value };
    setHeaders(updated);
  }

  function addParam() {
    setParams([
      ...params,
      { name: "", type: "string", description: "", required: false },
    ]);
  }

  function removeParam(index: number) {
    setParams(params.filter((_, i) => i !== index));
  }

  function updateParam(index: number, field: keyof ParamRow, value: any) {
    const updated = [...params];
    updated[index] = { ...updated[index], [field]: value };
    setParams(updated);
  }

  async function handleSubmit() {
    if (!name.trim()) {
      setError("Name is required");
      return;
    }
    if (!url.trim()) {
      setError("URL is required");
      return;
    }
    const urlErr = validateUrl(url);
    if (urlErr) {
      setUrlError(urlErr);
      return;
    }

    setSaving(true);
    setError("");

    const inputSchema = buildInputSchema(params);
    const validHeaders = headers.filter((h) => h.name.trim());

    try {
      if (isEdit) {
        const { error: err } = await fetchClient.PUT(
          "/api/v1/http-tools/{id}",
          {
            params: { path: { id: tool.id } },
            body: {
              name: name.trim(),
              description: description.trim(),
              method: method as any,
              url: url.trim(),
              headers: validHeaders,
              bodyTemplate,
              inputSchema,
            },
          },
        );
        if (err) throw new Error((err as any).error);
      } else {
        const { error: err } = await fetchClient.POST("/api/v1/http-tools", {
          body: {
            name: name.trim(),
            description: description.trim() || undefined,
            method: method as any,
            url: url.trim(),
            headers: validHeaders.length > 0 ? validHeaders : undefined,
            bodyTemplate: bodyTemplate || undefined,
            inputSchema: inputSchema !== "{}" ? inputSchema : undefined,
          },
        });
        if (err) throw new Error((err as any).error);
      }
      onSaved();
    } catch (err: any) {
      setError(err.message || "Failed to save HTTP tool");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Stack>
      <Button
        variant="subtle"
        leftSection={<IconArrowLeft size={16} />}
        onClick={onCancel}
        size="sm"
        w="fit-content"
      >
        Back to HTTP Tools
      </Button>

      <Text size="xl" fw={600}>
        {isEdit ? "Edit HTTP Tool" : "New HTTP Tool"}
      </Text>

      {error && (
        <Alert color="red" variant="light">
          {error}
        </Alert>
      )}

      <TextInput
        label="Name"
        description="The LLM uses this as the function name when calling the tool"
        placeholder="e.g. get_weather"
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
        required
      />

      <Textarea
        label="Description"
        description="The LLM uses this to decide when to call the tool"
        placeholder="e.g. Retrieves current weather data for a given city"
        resize="vertical"
        value={description}
        onChange={(e) => setDescription(e.currentTarget.value)}
        minRows={2}
      />

      <Card withBorder padding="md">
        <Text size="sm" fw={500} mb="xs">
          Input Parameters
        </Text>
        <Text size="xs" c="dimmed" mb="sm">
          Define the parameters the LLM must provide when calling this tool.
          These become template variables for URL, headers, and body.
        </Text>

        <Stack gap="xs">
          {params.map((param, i) => (
            <Group key={i} gap="xs" align="flex-end">
              <TextInput
                label={i === 0 ? "Name" : undefined}
                placeholder="param name"
                value={param.name}
                onChange={(e) => updateParam(i, "name", e.currentTarget.value)}
                style={{ flex: 2 }}
                size="sm"
              />
              <Select
                label={i === 0 ? "Type" : undefined}
                data={PARAM_TYPES}
                value={param.type}
                onChange={(v) => updateParam(i, "type", v ?? "string")}
                style={{ flex: 1 }}
                size="sm"
              />
              <TextInput
                label={i === 0 ? "Description" : undefined}
                placeholder="parameter description"
                value={param.description}
                onChange={(e) =>
                  updateParam(i, "description", e.currentTarget.value)
                }
                style={{ flex: 3 }}
                size="sm"
              />
              <Checkbox
                label="Req"
                checked={param.required}
                onChange={(e) =>
                  updateParam(i, "required", e.currentTarget.checked)
                }
                size="sm"
              />
              <ActionIcon
                variant="subtle"
                color="red"
                onClick={() => removeParam(i)}
                size="lg"
              >
                <IconTrash size={24} />
              </ActionIcon>
            </Group>
          ))}
        </Stack>

        <Button
          variant="light"
          size="xs"
          leftSection={<IconPlus size={16} />}
          onClick={addParam}
          mt="xs"
        >
          Add Parameter
        </Button>
      </Card>

      <Group gap="sm" grow>
        <Select
          label="Method"
          description="HTTP method to use when calling the tool"
          data={HTTP_METHODS}
          value={method}
          onChange={(v) => setMethod((v as HttpMethod) ?? "GET")}
          required
        />
        <TextInput
          label="URL"
          description="Supports Go template expressions, e.g. https://api.example.com/{{.city}}"
          placeholder="https://api.example.com/endpoint"
          value={url}
          onChange={(e) => {
            setUrl(e.currentTarget.value)
            if (urlError) setUrlError("")
          }}
          onBlur={(e) => setUrlError(validateUrl(e.currentTarget.value))}
          error={urlError}
          required
          style={{ flex: 3 }}
        />
      </Group>

      <Card withBorder padding="md">
        <Text size="sm" fw={500} mb="xs">
          Headers
        </Text>
        <Text size="xs" c="dimmed" mb="sm">
          Values support Go template expressions, e.g. {"Bearer {{.token}}"}
        </Text>

        <Stack gap="xs">
          {headers.map((header, i) => (
            <Group key={i} gap="xs" align="flex-end">
              <TextInput
                label={i === 0 ? "Name" : undefined}
                placeholder="Header-Name"
                value={header.name}
                onChange={(e) => updateHeader(i, "name", e.currentTarget.value)}
                style={{ flex: 1 }}
                size="sm"
              />
              <TextInput
                label={i === 0 ? "Value" : undefined}
                placeholder="Header value"
                value={header.value}
                onChange={(e) =>
                  updateHeader(i, "value", e.currentTarget.value)
                }
                style={{ flex: 2 }}
                size="sm"
              />
              <ActionIcon
                variant="subtle"
                color="red"
                onClick={() => removeHeader(i)}
                size="lg"
              >
                <IconTrash size={24} />
              </ActionIcon>
            </Group>
          ))}
        </Stack>

        <Button
          variant="light"
          size="xs"
          leftSection={<IconPlus size={16} />}
          onClick={addHeader}
          mt="xs"
        >
          Add Header
        </Button>
      </Card>

      <Stack gap={4}>
        <Group justify="space-between" align="flex-end">
          <div>
            <Text size="sm" fw={500}>
              Body Template
            </Text>
            <Text size="xs" c="dimmed">
              {
                'Supports Go template expressions, e.g. {"query": "{{.searchTerm}}"}'
              }
            </Text>
          </div>
          <Button
            variant="subtle"
            size="compact-xs"
            leftSection={<IconBraces size={14} />}
            onClick={() => {
              try {
                const placeholders: string[] = [];
                const stripped = bodyTemplate.replace(
                  /\{\{.*?\}\}/g,
                  (match) => {
                    placeholders.push(match);
                    return `"__TPL_${placeholders.length - 1}__"`;
                  },
                );
                const formatted = JSON.stringify(JSON.parse(stripped), null, 2);
                const restored = formatted.replace(
                  /"__TPL_(\d+)__"/g,
                  (_, idx) => placeholders[Number(idx)],
                );
                setBodyTemplate(restored);
              } catch {
                // Not valid JSON — try formatting as-is
                try {
                  setBodyTemplate(
                    JSON.stringify(JSON.parse(bodyTemplate), null, 2),
                  );
                } catch {
                  // Can't parse, leave unchanged
                }
              }
            }}
          >
            Beautify JSON
          </Button>
        </Group>
        <Textarea
          placeholder='{"key": "value"}'
          resize="vertical"
          value={bodyTemplate}
          onChange={(e) => setBodyTemplate(e.currentTarget.value)}
          minRows={4}
          styles={{ input: { fontFamily: "monospace", fontSize: "13px" } }}
        />
      </Stack>

      <Group>
        <Button onClick={handleSubmit} loading={saving}>
          {isEdit ? "Save Changes" : "Create HTTP Tool"}
        </Button>
        <Button variant="default" onClick={onCancel}>
          Cancel
        </Button>
      </Group>
    </Stack>
  );
}
