<script lang="ts">
  import Modal from '../../../shared/ui/Modal.svelte';
  import { t } from '../../../shared/i18n';
  import type { DeviceUseCases } from '../usecase/deviceUseCases';
  import DisplayModeChoice from './DisplayModeChoice.svelte';

  /** The setup step of an appliance (#140): asked once, until a mode is chosen. */
  interface Props {
    device: DeviceUseCases;
    onclose: () => void;
  }
  let { device, onclose }: Props = $props();
</script>

<Modal title={t('device.question')} {onclose}>
  <DisplayModeChoice {device} onchosen={onclose} />
  <p class="muted small">{t('device.later')}</p>
</Modal>

<style>
  .small {
    font-size: 0.85em;
  }
</style>
