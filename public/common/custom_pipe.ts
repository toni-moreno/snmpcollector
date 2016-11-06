import { Pipe, PipeTransform } from '@angular/core';

@Pipe({name: 'objectParser'})
export class ObjectParserPipe implements PipeTransform {
  transform(value) : any {
    let keys = [];
    for (let key in value) {
      keys.push({key: key, value: value[key]});
    }
    return keys;
}
}
