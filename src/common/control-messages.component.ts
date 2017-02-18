
//https://coryrylan.com/blog/angular-2-form-builder-and-validation-management

import {Component, Input} from '@angular/core';
import {FormGroup,FormControl} from '@angular/forms';
import {ValidationService} from './validation.service';

@Component({
    selector: 'control-messages',
    template: `
      <label [ngClass]="['label', 'label-'+errorMessage.alertType]" *ngIf="errorMessage !== null"><i class="glyphicon glyphicon-warning-sign"></i> {{errorMessage.message}}</label>`
})

export class ControlMessagesComponent {
  @Input() control: FormControl;
  constructor() { }

  get errorMessage() {
    let alertType = '';
    for (let propertyName in this.control.errors) {
      if (this.control.errors.hasOwnProperty(propertyName)) {
        this.control.errors.hasOwnProperty('required') ? alertType = 'danger' : alertType = 'warning';
        return {'message': ValidationService.getValidatorErrorMessage(propertyName, this.control.errors[propertyName]),'alertType': alertType};
      }
    }
    return null;
  }
}
