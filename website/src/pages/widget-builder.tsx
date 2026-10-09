import type {ReactNode} from 'react';
import {lazy, Suspense} from 'react';
import BrowserOnly from '@docusaurus/BrowserOnly';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';

// Loaded lazily and only in the browser: the builder injects the chat widget into the page.
const WidgetBuilderApp = lazy(() => import('../components/WidgetBuilderApp'));

export default function WidgetBuilderPage(): ReactNode {
  return (
    <Layout
      title="Chat widget builder"
      description="Design an AI chat widget visually and copy the embed code. Works with any backend, including GenKitKraft agents.">
      <main className="container margin-vert--lg">
        <h1>Chat widget builder</h1>
        <p>
          Design the <a href="https://github.com/techorionai/ai-chat-widget">AI chat widget</a>{' '}
          visually, preview it live (the launcher appears at the bottom right of this page) and copy the
          embed code. To connect it to a GenKitKraft agent, follow the{' '}
          <Link to="/docs/guides/chat-widget">integration guide</Link>.
        </p>
        <BrowserOnly fallback={<p>Loading builder…</p>}>
          {() => (
            <Suspense fallback={<p>Loading builder…</p>}>
              <WidgetBuilderApp />
            </Suspense>
          )}
        </BrowserOnly>
      </main>
    </Layout>
  );
}
