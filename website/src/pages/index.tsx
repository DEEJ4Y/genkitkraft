import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

type FeatureItem = {
  icon: string;
  title: string;
  description: string;
};

const FeatureList: FeatureItem[] = [
  {
    icon: '🤖',
    title: 'Agent Builder UI',
    description:
      'Create and configure AI agents with a visual interface. Set system prompts, choose models, and tune generation parameters.',
  },
  {
    icon: '🔌',
    title: 'Multi-Provider LLM Support',
    description:
      'Connect to OpenAI, Anthropic, Google AI, Vertex AI, xAI, and DeepSeek. Switch providers without changing your workflow.',
  },
  {
    icon: '🛠️',
    title: 'MCP Tool Support',
    description:
      'Extend agents with Model Context Protocol tool servers. Give your agents the ability to interact with external systems.',
  },
  {
    icon: '🔗',
    title: 'OpenAI-Compatible API',
    description:
      'Expose configured agents through a standard OpenAI-compatible chat completions endpoint for easy integration.',
  },
  {
    icon: '💬',
    title: 'Built-in Playground',
    description:
      'Test agents interactively with a streaming chat playground. Manage sessions and experiment with different configurations.',
  },
  {
    icon: '📦',
    title: 'Single Binary Deployment',
    description:
      'Server and UI ship as one binary. Deploy with Docker in minutes with SQLite storage and zero external dependencies.',
  },
];

function Feature({icon, title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4', styles.featureCol)}>
      <div className={styles.featureCard}>
        <div className={styles.featureIcon}>{icon}</div>
        <Heading as="h3" className={styles.featureTitle}>
          {title}
        </Heading>
        <p className={styles.featureDescription}>{description}</p>
      </div>
    </div>
  );
}

function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs">
            Get Started
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Self-hostable LLM agent platform built on Google Genkit">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
