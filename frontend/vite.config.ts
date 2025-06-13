import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    TanStackRouterVite({ target: 'react', autoCodeSplitting: true }),
    react(),
  ],
  resolve: {
    alias: {
      '@': '/src',
      "@styles": '/src/styles',
      "@components": '/src/components',
      "@routes": '/src/routes',
      "@utils": '/src/utils',
      "@hooks": '/src/hooks',
      "@assets": '/src/assets',
    },
  },
  build: {
		outDir: "../backend/dist",
	},
})