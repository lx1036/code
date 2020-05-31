

/**
 * Represents label object used in deploy form view.
 */
export class DeployLabel {
  editable: boolean;
  key: string;
  value: string;

  constructor(key = '', value = '', editable = true) {
    this.editable = editable;

    this.key = key;

    this.value = value;
  }
}
