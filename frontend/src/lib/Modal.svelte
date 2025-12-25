<!-- Reusable Modal Component -->
<script>
  import { createEventDispatcher } from 'svelte';
  
  export let show = false;
  export let title = '';
  export let message = '';
  export let icon = '⚠️';
  export let type = 'confirm'; // 'confirm', 'alert', 'danger'
  export let confirmText = 'Confirm';
  export let cancelText = 'Cancel';
  
  const dispatch = createEventDispatcher();
  
  function handleConfirm() {
    dispatch('confirm');
    show = false;
  }
  
  function handleCancel() {
    dispatch('cancel');
    show = false;
  }
  
  function handleOverlayClick() {
    if (type === 'alert') {
      handleConfirm();
    } else {
      handleCancel();
    }
  }
</script>

{#if show}
  <div class="modal-overlay" on:click={handleOverlayClick}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-icon">{icon}</div>
      <h2 class="modal-title">{title}</h2>
      {#if message}
        <p class="modal-message">{@html message}</p>
      {/if}
      <slot></slot>
      <div class="modal-actions">
        {#if type === 'alert'}
          <button class="btn btn-primary" on:click={handleConfirm}>
            {confirmText}
          </button>
        {:else}
          <button class="btn btn-secondary" on:click={handleCancel}>
            {cancelText}
          </button>
          <button 
            class="btn {type === 'danger' ? 'btn-danger' : 'btn-primary'}" 
            on:click={handleConfirm}
          >
            {confirmText}
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: var(--space-md);
    animation: fadeIn 0.15s ease;
  }
  
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  
  .modal {
    background: var(--surface);
    border-radius: var(--radius-card);
    padding: var(--space-lg);
    max-width: 340px;
    width: 100%;
    text-align: center;
    animation: slideUp 0.2s ease;
  }
  
  @keyframes slideUp {
    from { 
      opacity: 0;
      transform: translateY(20px);
    }
    to { 
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  .modal-icon {
    font-size: 48px;
    margin-bottom: var(--space-md);
  }
  
  .modal-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: var(--space-sm);
    color: var(--text-primary);
  }
  
  .modal-message {
    color: var(--text-secondary);
    font-size: 14px;
    margin-bottom: var(--space-lg);
    line-height: 1.5;
  }
  
  .modal-actions {
    display: flex;
    gap: var(--space-sm);
  }
  
  .modal-actions .btn {
    flex: 1;
  }
</style>
