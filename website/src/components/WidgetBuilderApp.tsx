import type {ReactNode} from 'react';
import {MantineProvider} from '@mantine/core';
import {useColorMode} from '@docusaurus/theme-common';
import useBaseUrl from '@docusaurus/useBaseUrl';
import {WidgetBuilder} from 'genkitkraft-widget-builder';
// Layered build: Mantine rules sit below unlayered CSS, so Infima (the docs theme) keeps priority
// on the rest of the site once this chunk has been loaded.
import '@mantine/core/styles.layer.css';

export default function WidgetBuilderApp(): ReactNode {
  const {colorMode} = useColorMode();
  const docsUrl = useBaseUrl('/docs/guides/chat-widget');
  return (
    <MantineProvider forceColorScheme={colorMode} withGlobalClasses={false}>
      <WidgetBuilder
        docsUrl={docsUrl}
        backendBaseUrl="/api/chat"
      />
    </MantineProvider>
  );
}
