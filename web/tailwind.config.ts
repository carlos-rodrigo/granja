import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./hooks/**/*.{js,ts,jsx,tsx,mdx}",
    "./lib/**/*.{js,ts,jsx,tsx,mdx}"
  ],
  theme: {
    extend: {
      colors: {
        app: {
          bg: "#090b10",
          panel: "#111722",
          panelAlt: "#161d2b",
          border: "#233047",
          text: "#e7eefc",
          muted: "#9fb0ca"
        },
        status: {
          planted: "#4f84ff",
          growing: "#f6c64a",
          ready: "#34d399",
          harvested: "#a78bfa",
          blocked: "#f87171"
        }
      },
      boxShadow: {
        glow: "0 12px 48px rgba(16, 28, 54, 0.35)"
      }
    }
  },
  plugins: []
};

export default config;
