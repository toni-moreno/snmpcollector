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

@Pipe({name: 'splitComma'})
export class SplitCommaPipe implements PipeTransform {
  transform(value) : any {
    let valArray = [];
    valArray = value.split(',');
    console.log(valArray);
    return valArray;
}
}
