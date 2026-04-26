package sqliteagenttool

import (
	"context"
	"database/sql"

	agenttoolrepo "github.com/DEEJ4Y/genkitkraft/internal/ports/agent_tool_repo"
)

var _ agenttoolrepo.AgentToolRepository = (*Repository)(nil)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByAgentID(ctx context.Context, agentID string) (agenttoolrepo.AgentToolConfig, error) {
	config := agenttoolrepo.AgentToolConfig{
		AgentID:     agentID,
		HttpToolIDs: []string{},
		McpServers:  []agenttoolrepo.McpServerToolConfig{},
	}

	// Load HTTP tool IDs
	rows, err := r.db.QueryContext(ctx, `SELECT http_tool_id FROM agent_http_tools WHERE agent_id = ?`, agentID)
	if err != nil {
		return config, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return config, err
		}
		config.HttpToolIDs = append(config.HttpToolIDs, id)
	}
	if err := rows.Err(); err != nil {
		return config, err
	}

	// Load MCP server configs
	serverRows, err := r.db.QueryContext(ctx, `SELECT mcp_server_id, select_all FROM agent_mcp_servers WHERE agent_id = ?`, agentID)
	if err != nil {
		return config, err
	}
	defer serverRows.Close()

	type serverRow struct {
		id        string
		selectAll bool
	}
	var servers []serverRow
	for serverRows.Next() {
		var s serverRow
		if err := serverRows.Scan(&s.id, &s.selectAll); err != nil {
			return config, err
		}
		servers = append(servers, s)
	}
	if err := serverRows.Err(); err != nil {
		return config, err
	}

	for _, s := range servers {
		mc := agenttoolrepo.McpServerToolConfig{
			McpServerID: s.id,
			SelectAll:   s.selectAll,
			ToolNames:   []string{},
		}

		if !s.selectAll {
			toolRows, err := r.db.QueryContext(ctx,
				`SELECT tool_name FROM agent_mcp_server_tools WHERE agent_id = ? AND mcp_server_id = ?`,
				agentID, s.id)
			if err != nil {
				return config, err
			}
			for toolRows.Next() {
				var name string
				if err := toolRows.Scan(&name); err != nil {
					toolRows.Close()
					return config, err
				}
				mc.ToolNames = append(mc.ToolNames, name)
			}
			toolRows.Close()
			if err := toolRows.Err(); err != nil {
				return config, err
			}
		}

		config.McpServers = append(config.McpServers, mc)
	}

	return config, nil
}

func (r *Repository) Save(ctx context.Context, config agenttoolrepo.AgentToolConfig) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear existing
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_mcp_server_tools WHERE agent_id = ?`, config.AgentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_mcp_servers WHERE agent_id = ?`, config.AgentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM agent_http_tools WHERE agent_id = ?`, config.AgentID); err != nil {
		return err
	}

	// Insert HTTP tools
	for _, toolID := range config.HttpToolIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO agent_http_tools (agent_id, http_tool_id) VALUES (?, ?)`,
			config.AgentID, toolID); err != nil {
			return err
		}
	}

	// Insert MCP servers and their tools
	for _, mc := range config.McpServers {
		selectAllInt := 0
		if mc.SelectAll {
			selectAllInt = 1
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO agent_mcp_servers (agent_id, mcp_server_id, select_all) VALUES (?, ?, ?)`,
			config.AgentID, mc.McpServerID, selectAllInt); err != nil {
			return err
		}

		if !mc.SelectAll {
			for _, toolName := range mc.ToolNames {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO agent_mcp_server_tools (agent_id, mcp_server_id, tool_name) VALUES (?, ?, ?)`,
					config.AgentID, mc.McpServerID, toolName); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
