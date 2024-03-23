import { defineConfig, transformWithEsbuild } from 'vite';
import react from '@vitejs/plugin-react';
import { splitVendorChunkPlugin } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    {
      name: 'treat-js-files-as-jsx',
      async transform(code, id) {
        if (!id.match(/src\/.*\.js$/))  return null

        // Use the exposed transform from vite, instead of directly
        // transforming with esbuild
        return transformWithEsbuild(code, id, {
          loader: 'jsx',
          jsx: 'automatic',
        })
      },
    },
    react(),
    splitVendorChunkPlugin()
  ],
  optimizeDeps: {
    force: true,
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'semi': ['@douyinfe/semi-ui'],
          'icons': ['@douyinfe/semi-icons'],
          'semantic': ['semantic-ui-react'],
          'visactor': ['@visactor/react-vchart', '@visactor/vchart']
        },
      },
    },
  },
  server: {
    proxy: {
      '/api': {
        target: "http://localhost:3000",
        changeOrigin: true,
      }
    }
  }
});
