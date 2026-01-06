import { createRootRoute, Link, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/router-devtools'

export const Route = createRootRoute({
  component: RootComponent,
})

function RootComponent() {
  return (
    <html lang="en">
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Podoru Chain Explorer</title>
        <style>{`
          * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
          }

          body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background-color: #f5f7fa;
            color: #2d3748;
            line-height: 1.6;
          }

          .header {
            background-color: #2d3748;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
          }

          .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
          }

          .nav {
            display: flex;
            gap: 1.5rem;
            margin-top: 1rem;
          }

          .nav a {
            color: #a0aec0;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            transition: all 0.2s;
          }

          .nav a:hover,
          .nav a.active {
            color: white;
            background-color: #4a5568;
          }

          .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 2rem;
          }

          .card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            margin-bottom: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
          }

          .card h2 {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            color: #2d3748;
          }

          .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
          }

          .status {
            display: inline-block;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.875rem;
            font-weight: 500;
          }

          .status.connected {
            background-color: #c6f6d5;
            color: #22543d;
          }

          .status.disconnected {
            background-color: #fed7d7;
            color: #742a2a;
          }

          .hash {
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
            color: #4a5568;
          }

          .table {
            width: 100%;
            border-collapse: collapse;
          }

          .table th,
          .table td {
            text-align: left;
            padding: 0.75rem;
            border-bottom: 1px solid #e2e8f0;
          }

          .table th {
            font-weight: 600;
            color: #4a5568;
            background-color: #f7fafc;
          }

          .table tr:hover {
            background-color: #f7fafc;
          }

          button {
            background-color: #4299e1;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.875rem;
            transition: background-color 0.2s;
          }

          button:hover {
            background-color: #3182ce;
          }

          input {
            width: 100%;
            padding: 0.75rem;
            border: 1px solid #cbd5e0;
            border-radius: 4px;
            font-size: 1rem;
          }

          input:focus {
            outline: none;
            border-color: #4299e1;
            box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
          }
        `}</style>
      </head>
      <body>
        <div className="header">
          <h1>⛓️ Podoru Chain Explorer</h1>
          <nav className="nav">
            <Link to="/" activeOptions={{ exact: true }}>Dashboard</Link>
            <Link to="/blocks">Blocks</Link>
            <Link to="/transactions">Transactions</Link>
            <Link to="/state">State Browser</Link>
            <Link to="/network">Network</Link>
          </nav>
        </div>
        <div className="container">
          <Outlet />
        </div>
        <TanStackRouterDevtools position="bottom-right" />
      </body>
    </html>
  )
}
