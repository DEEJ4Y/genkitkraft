namespace Api.Models;

/** A built-in tool available in GenKitKraft. */
model BuiltInToolResponse {
  /** Unique identifier for the built-in tool. */
  id: string;

  /** Human-readable name of the tool. */
  name: string;

  /** Description of what the tool does. */
  description: string;
}

/** Response containing the list of available built-in tools. */
model BuiltInToolListResponse {
  /** List of available built-in tools. */
  builtInTools: BuiltInToolResponse[];
}
