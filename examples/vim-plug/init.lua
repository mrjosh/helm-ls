-- a minimal example config for setting up neovim with helm-ls and yamlls using the plugin manager vim-plug
-- does not cover syntax highlighting, see ../nvim/init.lua for that
-- test it with: nvim -u init.lua
-- make sure to follow the vim-plug installation instructions: https://github.com/junegunn/vim-plug

local Plug = vim.fn["plug#"]

-- install required plugins
vim.call("plug#begin")

Plug("https://github.com/neovim/nvim-lspconfig")
Plug("https://github.com/qvalentin/helm-ls.nvim")

vim.call("plug#end")

-- setup helm-ls.nvim
require("helm-ls").setup({
  conceal_templates = {
    -- enable the replacement of templates with virtual text of their current values
    enabled = false, -- tree-sitter must be setup for this feature
  },
  indent_hints = {
    -- enable hints for indent and nindent functions
    enabled = false, -- tree-sitter must be setup for this feature
  },
})

local lspconfig = require("lspconfig")

lspconfig.yamlls.setup({})
-- setup helm-ls
lspconfig.helm_ls.setup({
  settings = {
    ["helm-ls"] = {
      yamlls = {
        path = "yaml-language-server",
      },
    },
  },
})

-- below is the config for a basic lsp keymap
vim.keymap.set("n", "<space>e", vim.diagnostic.open_float)
vim.keymap.set("n", "<space>q", vim.diagnostic.setloclist)

-- Use LspAttach autocommand to only map the following keys
-- after the language server attaches to the current buffer
vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("UserLspConfig", {}),
  callback = function(ev)
    -- Enable completion triggered by <c-x><c-o>
    vim.bo[ev.buf].omnifunc = "v:lua.vim.lsp.omnifunc"

    -- Buffer local mappings.
    -- See `:help vim.lsp.*` for documentation on any of the below functions
    local opts = { buffer = ev.buf }
    vim.keymap.set("n", "gD", vim.lsp.buf.declaration, opts)
    vim.keymap.set("n", "gd", vim.lsp.buf.definition, opts)
    vim.keymap.set("n", "K", vim.lsp.buf.hover, opts)
    vim.keymap.set("n", "gi", vim.lsp.buf.implementation, opts)
    vim.keymap.set("n", "<C-k>", vim.lsp.buf.signature_help, opts)
    vim.keymap.set("n", "<space>wa", vim.lsp.buf.add_workspace_folder, opts)
    vim.keymap.set("n", "<space>wr", vim.lsp.buf.remove_workspace_folder, opts)
    vim.keymap.set("n", "<space>wl", function()
      print(vim.inspect(vim.lsp.buf.list_workspace_folders()))
    end, opts)
    vim.keymap.set("n", "<space>D", vim.lsp.buf.type_definition, opts)
    vim.keymap.set("n", "<space>rn", vim.lsp.buf.rename, opts)
    vim.keymap.set({ "n", "v" }, "<space>ca", vim.lsp.buf.code_action, opts)
    vim.keymap.set("n", "gr", vim.lsp.buf.references, opts)
    vim.keymap.set("n", "<space>f", function()
      vim.lsp.buf.format({ async = true })
    end, opts)
  end,
})
