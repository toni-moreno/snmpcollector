
//https://coryrylan.com/blog/angular-2-form-builder-and-validation-management

import {Component, Input} from '@angular/core';
import {FormGroup,FormControl} from '@angular/forms';
import {ValidationService} from './validation.service';

@Component({
    selector: 'control-messages',
    template: `<span class="bg-danger" *ngIf="errorMessage !== null">{{errorMessage}}</span>`
})

export class ControlMessagesComponent {
  @Input() control: FormControl;
  constructor() { }

  get errorMessage() {
    for (let propertyName in this.control.errors) {
      if (this.control.errors.hasOwnProperty(propertyName) && this.control.touched) {
        return ValidationService.getValidatorErrorMessage(propertyName, this.control.errors[propertyName]);
      }
    }

    return null;
  }
}
