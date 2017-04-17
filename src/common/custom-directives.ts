import { Directive, ElementRef, HostListener, Input } from '@angular/core';
@Directive({
  selector: '[passwordToggle]'
})
export class PasswordToggleDirective {
  @Input() input: any;
  inputtype : string = 'password';
  constructor() {}

  @HostListener('mouseup') onMouseEnter() {
    this.inputtype === 'password' ? this.inputtype = 'text' : this.inputtype = 'password';
    this.toggleType(this.inputtype);
  }
  private toggleType (value: string) {
    this.input.type = value;
  }
}
