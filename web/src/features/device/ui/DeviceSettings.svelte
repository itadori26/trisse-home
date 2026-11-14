<script lang="ts">
  import { t } from '../../../shared/i18n';
  import type { DeviceUseCases } from '../usecase/deviceUseCases';
  import DisplayModeChoice from './DisplayModeChoice.svelte';

  /** Settings → Device (#140, appliance only). */
  interface Props {
    device: DeviceUseCases;
  }
  let { device }: Props = $props();
</script>

<p class="muted">{t('device.settings_hint')}</p>
<DisplayModeChoice {device} />
