import { defineConfig } from 'tsup'

export default defineConfig({
  entry: ['src/index.ts'],
  format: ['esm'],
  dts: true,
  clean: true,
  sourcemap: false,
  external: ['react', 'react-dom', 'react/jsx-runtime', '@mantine/core', '@mantine/hooks', '@tabler/icons-react', 'navigableai-chat-widget'],
})
